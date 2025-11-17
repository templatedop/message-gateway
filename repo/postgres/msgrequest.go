package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"MgApplication/core/domain"

	config "MgApplication/api-config"
	dblib "MgApplication/api-db"
	log "MgApplication/api-log"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
)

type MgApplicationRepository struct {
	Db  *dblib.DB
	Cfg *config.Config
}

// NewOfficeRepository creates a new Office repository instance
func NewMgApplicationRepository(Db *dblib.DB, Cfg *config.Config) *MgApplicationRepository {
	return &MgApplicationRepository{
		Db,
		Cfg,
	}
}
func CallAPI(url string, method string, headers map[string]string, params map[string]interface{}) (map[string]interface{}, error) {

	// fmt.Print(params)
	// tr := &http.Transport{
	// 	TLSClientConfig: &tls.Config{
	// 		MinVersion:         tls.VersionTLS12,
	// 		InsecureSkipVerify: false,                       // Set to true only for testing; not recommended for production.
	// 		Renegotiation:      tls.RenegotiateOnceAsClient, // Adjust renegotiation settings.
	// 	},
	// 	DisableKeepAlives: true, // Disable keep-alive
	// }

	client := resty.New().SetTimeout(30 * time.Second)
	// client.SetTransport(tr)
	request := client.R()
	request.SetHeaders(headers)

	switch method {
	case "GET":
		stringParams := ConvertMapToStringMap(params)
		request.SetQueryParams(stringParams)
	case "POST", "PUT", "DELETE":
		request.SetBody(params)
		//stringParams := ConvertMapToStringMap(params)
		//request.SetFormData(stringParams)
	default:
		return nil, errors.New("unsupported HTTP method")
	}

	response, err := request.Execute(method, url)
	if err != nil {
		return nil, err
	}

	// fmt.Print(response)
	var responseBody map[string]interface{}
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		return nil, err
	}
	if errorCode, exists := responseBody["error_code"]; exists {
		errorMessage := responseBody["message"].(string)
		fmt.Printf("Error: Code %v - %v\n", errorCode, errorMessage)
		return map[string]interface{}{}, errors.New(errorMessage)
	}

	return responseBody, nil
}
func ConvertMapToStringMap(params map[string]interface{}) map[string]string {
	stringParams := make(map[string]string)
	for key, value := range params {
		stringParams[key] = interfaceToString(value)
	}
	return stringParams
}
func interfaceToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	// Add cases for other types as needed
	default:
		return fmt.Sprintf("%v", v)
	}
}
func (cr *MgApplicationRepository) SendMsgToKafka(gctx *context.Context, url string, schema string, msgreq *domain.MsgRequest) (map[string]interface{}, error) {
	fmt.Println("kafka url is:", url)
	fmt.Println("kafka schema is:", schema)
	// Define Headers
	headers := map[string]string{
		"Content-Type": "application/vnd.kafka.avro.v2+json",
		"Accept":       "application/vnd.kafka.v2+json",
	}
	schemaint64, err := strconv.Atoi(schema)
	if err != nil {
		fmt.Println("Error:", err)
		return map[string]interface{}{}, err
	}
	// Define Payload
	params := map[string]interface{}{
		"value_schema_id": schemaint64,
		"records": []map[string]interface{}{
			{
				"value": map[string]interface{}{
					"reqid":          msgreq.RequestID,
					"application_id": msgreq.ApplicationID,
					"facility_id":    msgreq.FacilityID,
					"priority":       msgreq.Priority,
					"message_text":   msgreq.MessageText,
					"sender_id":      msgreq.SenderID,
					"mobile_numbers": msgreq.MobileNumbers,
					"entity_id":      msgreq.EntityId,
					"template_id":    msgreq.TemplateID,
					"message_type":   msgreq.MessageType,
				},
			},
		},
	}

	// Call the API
	response, err := CallAPI(url, "POST", headers, params)
	if err != nil {
		fmt.Println("Error calling API:", err)
		return map[string]interface{}{}, err
	}
	fmt.Println("Response from callAPI:", response)
	return response, nil
}
func (cr *MgApplicationRepository) SaveMsgRequestTx(gctx *context.Context, msgapp *domain.MsgRequest) (*domain.MsgRequest, error) {

	ctx, cancel := context.WithTimeout(context.Background(), cr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	log.Debug(nil, "Inside SaveMsgRequest Repo function")

	var Counter domain.Counter
	var msgreq1 domain.MsgRequest

	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		//checking whether applicaiton exists in the database
		query1 := psql.Select(
			sm.Columns("COUNT(1) as count"),
			sm.From("msg_application"),
			sm.Where(psql.Quote("application_id").EQ(psql.Arg(msgapp.ApplicationID))),
		)
		err := dblib.TxReturnRow(ctx, tx, query1, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error checking existence of application in msg_application table in SaveMsgRequest: %s", err.Error())
			return err
		}
		if Counter.Count == 0 {
			return errors.New("application does not exists")
		}
		//checking whether application and templateid combination available or not
		// query2 := dblib.Psql.Select("COUNT(1) as count").
		// 	From("msg_template").
		// 	Where(squirrel.And{squirrel.Eq{"ANY(string_to_array(application_id,','))": msgapp.ApplicationID}, squirrel.Eq{"template_id": msgapp.TemplateID}})

		query2 := psql.Select(
			sm.Columns("COUNT(1) AS count"),
			sm.From("msg_template"),
			sm.Where(psql.Raw("EXISTS (SELECT 1 FROM unnest(string_to_array(application_id, ',')) AS app_id WHERE app_id = ?)", msgapp.ApplicationID)),
			sm.Where(psql.Quote("template_id").EQ(psql.Arg(msgapp.TemplateID))),
		)
		err = dblib.TxReturnRow(ctx, tx, query2, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error checking whether a template registered for an application in SaveMsgRequest function: %s", err.Error())
			return err
		}
		if Counter.Count == 0 {
			return errors.New("application and template are not mapped. Contact CEPT")
		}
		numbers := strings.Split(msgapp.MobileNumbers, ",")
		var mobileNumbers []int64
		for _, numStr := range numbers {
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				log.Error(ctx, "Error converting %s to int64: %v\n", numStr, err)
				continue
			}
			mobileNumbers = append(mobileNumbers, num)
		}
		// Check if data already exists
		// Insert into msg_request and retrieve the gateway
		subquery := psql.Select(
			sm.Columns("mt.gateway"),
			sm.Columns(psql.Arg(msgapp.ApplicationID).As("application_id")),
			sm.Columns(psql.Arg(msgapp.FacilityID).As("facility_id")),
			sm.Columns(psql.Arg(msgapp.MessageText).As("message_text")),
			sm.Columns(psql.Arg(msgapp.SenderID).As("sender_id")),
			sm.Columns(psql.Arg(msgapp.EntityId).As("entity_id")),
			sm.Columns(psql.Arg(msgapp.TemplateID).As("template_id")),
			sm.Columns(psql.Arg("pending").As("status")),
			sm.Columns(psql.Arg(msgapp.Priority).As("priority")),
			sm.Columns(psql.Arg(mobileNumbers).As("mobile_number")),
			sm.From("msg_template mt"),
			sm.Where(psql.Quote("mt.template_id").EQ(psql.Arg(msgapp.TemplateID))),
		)

		query3 := psql.Insert(
			im.Into("msg_request", "gateway", "application_id", "facility_id", "message_text", "sender_id", "entity_id", "template_id", "status", "priority", "mobile_number"),
			im.Query(subquery),
			im.Returning("request_id", "communication_id", "gateway"),
		)

		msgreq1, err = dblib.InsertReturning(ctx, cr.Db, query3, pgx.RowToStructByNameLax[domain.MsgRequest])
		if err != nil {
			log.Error(ctx, "error executing insert query in SaveMsgRequest repo function: %w", err)
			return err
		}

		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Transaction rolling back in SaveMsgRequest repo function:  %s", TxDB.Error())
		return &domain.MsgRequest{}, TxDB
	}
	msgapp.Gateway = msgreq1.Gateway
	msgapp.CommunicationID = msgreq1.CommunicationID
	msgapp.RequestID = msgreq1.RequestID
	return msgapp, nil
}

func (cr *MgApplicationRepository) SaveMsgRequest(gctx *context.Context, msgapp *domain.MsgRequest) (*domain.MsgRequest, error) {

	ctx, cancel := context.WithTimeout(context.Background(), cr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	log.Debug(nil, "Inside SaveMsgRequest Repo function")

	var Counter domain.Counter
	var msgreq1 domain.MsgRequest

	//checking whether application exists in the database
	query1 := psql.Select(
		sm.Columns("COUNT(1) as count"),
		sm.From("msg_application"),
		sm.Where(psql.Quote("application_id").EQ(psql.Arg(msgapp.ApplicationID))),
	)
	Counter, err := dblib.SelectOne(ctx, cr.Db, query1, pgx.RowToStructByNameLax[domain.Counter])
	// err := dblib.ReturnRow(ctx, cr.Db, query1, pgx.RowToStructByNameLax[domain.Counter], &Counter)
	if err != nil {
		log.Error(ctx, "Error checking existence of application in msg_application table in SaveMsgRequest: %s", err.Error())
		return &domain.MsgRequest{}, err
	}
	if Counter.Count == 0 {
		return &domain.MsgRequest{}, errors.New("application does not exists")
	}

	//checking whether application and templateid combination available or not
	query2 := psql.Select(
		sm.Columns("COUNT(1) AS count"),
		sm.From("msg_template"),
		sm.Where(psql.Raw("EXISTS (SELECT 1 FROM unnest(string_to_array(application_id, ',')) AS app_id WHERE app_id = ?)", msgapp.ApplicationID)),
		sm.Where(psql.Quote("template_id").EQ(psql.Arg(msgapp.TemplateID))),
	)
	// err = dblib.ReturnRow(ctx, cr.Db, query2, pgx.RowToStructByNameLax[domain.Counter], &Counter)
	Counter, err = dblib.SelectOne(ctx, cr.Db, query2, pgx.RowToStructByNameLax[domain.Counter])
	if err != nil {
		log.Error(ctx, "Error checking whether a template registered for an application in SaveMsgRequest function: %s", err.Error())
		return &domain.MsgRequest{}, err
	}
	if Counter.Count == 0 {
		return &domain.MsgRequest{}, errors.New("application and template are not mapped. Contact CEPT")
	}

	numbers := strings.Split(msgapp.MobileNumbers, ",")
	var mobileNumbers []int64
	for _, numStr := range numbers {
		num, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			log.Error(ctx, "Error converting %s to int64: %v\n", numStr, err)
			continue
		}
		mobileNumbers = append(mobileNumbers, num)
	}

	// Insert into msg_request and retrieve the gateway
	subquery := psql.Select(
		sm.Columns("mt.gateway"),
		sm.Columns(psql.Arg(msgapp.ApplicationID).As("application_id")),
		sm.Columns(psql.Arg(msgapp.FacilityID).As("facility_id")),
		sm.Columns(psql.Arg(msgapp.MessageText).As("message_text")),
		sm.Columns(psql.Arg(msgapp.SenderID).As("sender_id")),
		sm.Columns(psql.Arg(msgapp.EntityId).As("entity_id")),
		sm.Columns(psql.Arg(msgapp.TemplateID).As("template_id")),
		sm.Columns(psql.Arg("pending").As("status")),
		sm.Columns(psql.Arg(msgapp.Priority).As("priority")),
		sm.Columns(psql.Arg(mobileNumbers).As("mobile_number")),
		sm.From("msg_template mt"),
		sm.Where(psql.Quote("mt.template_id").EQ(psql.Arg(msgapp.TemplateID))),
	)

	query3 := psql.Insert(
		im.Into("msg_request", "gateway", "application_id", "facility_id", "message_text", "sender_id", "entity_id", "template_id", "status", "priority", "mobile_number"),
		im.Query(subquery),
		im.Returning("request_id", "communication_id", "gateway"),
	)

	msgreq1, err = dblib.InsertReturning(ctx, cr.Db, query3, pgx.RowToStructByNameLax[domain.MsgRequest])
	if err != nil {
		log.Error(ctx, "error executing insert query in SaveMsgRequest repo function: %w", err)
		return &domain.MsgRequest{}, err
	}
	msgapp.Gateway = msgreq1.Gateway
	msgapp.CommunicationID = msgreq1.CommunicationID
	msgapp.RequestID = msgreq1.RequestID
	return msgapp, nil
}

func (cr *MgApplicationRepository) GetGateway(gctx *context.Context, msgreq *domain.MsgRequest) (*domain.MsgRequest, error) {

	ctx, cancel := context.WithTimeout(context.Background(), cr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	var msgreq1 domain.MsgRequest
	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query1 := psql.Select(
			sm.Columns("COUNT(1) as count"),
			sm.From("msg_template"),
			sm.Where(psql.Quote("template_id").EQ(psql.Arg(msgreq.TemplateID))),
		)
		err := dblib.TxReturnRow(ctx, tx, query1, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error checking whether a template exists or not in GetGateway repo function:  %s", err.Error())
			return err
		}
		if Counter.Count == 0 {
			return errors.New("template does not exists, hence cannot continue")
		}
		query2 := psql.Select(
			sm.Columns("0 as req_id", "'Not Applicable' as communication_id", "gateway", "entity_id", "message_type"),
			sm.From("msg_template"),
			sm.Where(psql.Quote("template_id").EQ(psql.Arg(msgreq.TemplateID))),
		)
		err = dblib.TxReturnRow(ctx, tx, query2, pgx.RowToStructByNameLax[domain.MsgRequest], &msgreq1)
		if err != nil {
			log.Error(ctx, "Error executing query in GetGateway repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Transaction rolling back in GetGateway repo function:  %s", TxDB.Error())
		return &domain.MsgRequest{}, TxDB
	}
	msgreq.RequestID = msgreq1.RequestID
	msgreq.CommunicationID = msgreq1.CommunicationID
	msgreq.Gateway = msgreq1.Gateway
	msgreq.EntityId = msgreq1.EntityId
	msgreq.MessageType = msgreq1.MessageType
	return msgreq, nil
}

func (cr *MgApplicationRepository) SaveGatewayDetailsTx(gctx *gin.Context, Gateway string, CommunicationID string) (bool, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), cr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := psql.Update(
			um.Table("msg_request"),
			um.Set("gateway", "updated_date").ToArg(Gateway, psql.Raw("current_timestamp")),
			um.Where(psql.Quote("communication_id").EQ(psql.Arg(CommunicationID))),
		)

		err := dblib.TxExec(ctx, tx, query)
		if err != nil {
			log.Error(ctx, "Error executing update query in SaveGatewayDetails repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Error initiating transaction in SaveGatewayDetails repo function:  %s", TxDB.Error())
		return false, TxDB
	}
	return true, nil
}

func (cr *MgApplicationRepository) SaveGatewayDetails(gctx *gin.Context, Gateway string, CommunicationID string) (bool, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), cr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()
	query := psql.Update(
		um.Table("msg_request"),
		um.Set("gateway", "updated_date").ToArg(Gateway, psql.Raw("current_timestamp")),
		um.Where(psql.Quote("communication_id").EQ(psql.Arg(CommunicationID))),
	)

	_, err := dblib.Update(ctx, cr.Db, query)
	if err != nil {
		log.Error(ctx, "Error executing update query in SaveGatewayDetails repo function:  %s", err.Error())
		return false, err
	}
	return true, nil
}

func (cr *MgApplicationRepository) SaveResponseTx(gctx *context.Context, msgRsp *domain.MsgResponse) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), cr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := psql.Update(
			um.Table("msg_request"),
			um.Set("status", "updated_date", "reference_id", "response_code", "response_message", "complete_response").ToArg("submitted", psql.Raw("current_timestamp"), msgRsp.ReferenceID, msgRsp.ResponseCode, msgRsp.ResponseText, msgRsp.CompleteResponse),
			um.Where(psql.Quote("communication_id").EQ(psql.Arg(msgRsp.CommunicationID))),
		)
		err := dblib.TxExec(ctx, tx, query)
		if err != nil {
			log.Error(ctx, "Error executing update query in SaveResponse repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Error initiating transaction in SaveResponse repo function:  %s", TxDB.Error())
		return false, TxDB
	}
	return true, nil
}

func (cr *MgApplicationRepository) SaveResponse(gctx *context.Context, msgRsp *domain.MsgResponse) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), cr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	query := psql.Update(
		um.Table("msg_request"),
		um.Set("status", "updated_date", "reference_id", "response_code", "response_message", "complete_response").ToArg("submitted", psql.Raw("current_timestamp"), msgRsp.ReferenceID, msgRsp.ResponseCode, msgRsp.ResponseText, msgRsp.CompleteResponse),
		um.Where(psql.Quote("communication_id").EQ(psql.Arg(msgRsp.CommunicationID))),
	)

	_, err := dblib.Update(ctx, cr.Db, query)
	if err != nil {
		log.Error(ctx, "Error executing update query in SaveResponse repo function:  %s", err.Error())
		return false, err
	}
	return true, nil
}
