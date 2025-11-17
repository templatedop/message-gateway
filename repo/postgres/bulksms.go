package repository

import (
	"MgApplication/core/domain"
	"context"
	"errors"
	"strconv"
	"strings"

	dblib "MgApplication/api-db"
	log "MgApplication/api-log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
)

func (cr *MgApplicationRepository) InitiateBulkSMSRepo(gctx *gin.Context, mbulk *domain.InitiateBulkSMS) (string, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), cr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	var mbulk1 domain.InitiateBulkSMS
	// var BulkSMS domain.InitiateBulkSMS
	var Template_Id, Sender_Id, Entity_Id, Message_Type string
	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		// Check if data already exists
		query1 := psql.Select(
			sm.Columns("COUNT(1) as count"),
			sm.From("msg_template"),
			sm.Where(psql.Quote("template_local_id").EQ(psql.Arg(mbulk.TemplateName))),
			sm.Where(psql.Raw("? = ANY(string_to_array(application_id, ','))", mbulk.ApplicationID)),
		)
		err := dblib.TxReturnRow(ctx, tx, query1, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error checking whether a msg_template exists for the given template_name and application_id in InitiateBulkSMS repo function:  %s", err.Error())
			return err
		}
		if Counter.Count == 0 {
			return errors.New("application and template are not mapped, refer maintain template")
		}
		query2 := psql.Select(
			sm.Columns("template_id", "entity_id", "sender_id", "message_type"),
			sm.From("msg_template"),
			sm.Where(psql.Quote("template_local_id").EQ(psql.Arg(mbulk.TemplateName))),
		)
		err = dblib.TxReturnRow(ctx, tx, query2, pgx.RowToStructByNameLax[domain.InitiateBulkSMS], &mbulk1)
		if err != nil {
			log.Error(ctx, "Error executing query in InitiateBulkSMS while collecting template_id, entity_id,sender_id and message_type: ", err)
			return err
		}
		Template_Id = mbulk1.TemplateID
		Sender_Id = mbulk1.SenderID
		Entity_Id = mbulk1.EntityID
		Message_Type = mbulk1.MessageType
		numbers := strings.Split(mbulk.MobileNo, ",")
		var mobileNumbers []int64
		for _, numStr := range numbers {
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				log.Error(ctx, "Error converting %s to int64: %v\n", numStr, err)
				continue
			}
			mobileNumbers = append(mobileNumbers, num)
		}
		query3 := psql.Insert(
			im.Into("msg_bulk_file", "application_id", "template_name", "template_id", "entity_id", "sender_id", "message_type", "mobile_number", "test_msg", "is_verified"),
			im.Values(psql.Arg(mbulk.ApplicationID, mbulk.TemplateName, mbulk1.TemplateID, mbulk1.EntityID, mbulk1.SenderID, mbulk1.MessageType, mobileNumbers, mbulk.TestMessage, false)),
			im.Returning("reference_id"),
		)
		err = dblib.TxReturnRow(ctx, tx, query3, pgx.RowToStructByNameLax[domain.InitiateBulkSMS], &mbulk1)
		if err != nil {
			log.Error(ctx, "Error executing insert query in InitiateBulkSMS repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Rolling back of transactions in InitiateBulkSMS repo function:  %s", TxDB.Error())
		return "", TxDB
	}
	return Template_Id + "_" + Sender_Id + "_" + Entity_Id + "_" + Message_Type + "_" + mbulk1.ReferenceID, nil
}

func (cr *MgApplicationRepository) ValidateTestSMSRepo(gctx *gin.Context, mbulk *domain.ValidateTestSMS) (bool, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), cr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := psql.Select(
			sm.Columns("COUNT(1) as count"),
			sm.From("msg_bulk_file"),
			sm.Where(psql.Quote("reference_id").EQ(psql.Arg(mbulk.ReferenceID))),
			sm.Where(psql.Quote("test_msg").Like(psql.Arg("%"+mbulk.TestString+"%"))),
		)
		err := dblib.TxReturnRow(ctx, tx, query, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error executing query in ValidateTestSMS repo function:  %s", err.Error())
			return err
		}
		if Counter.Count == 0 {
			//return errors.New("invalid test string, please refer the message sent to the mobile and enter one of the test number sent")
			return errors.New("Invalid test string, please refer to the message sent to the mobile and enter one of the test numbers sent")
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Rolling Back transaction in ValidateTestSMS repo function:  %s", TxDB.Error())
		return false, TxDB
	}
	return true, nil
}

/*
func (cr *MgApplicationRepository) GetTemplateDetails(gctx *gin.Context, msgtemplate *domain.MaintainTemplate) ([]domain.GetTemplateformatbyID, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), cr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var listTemplates []domain.GetTemplateformatbyID

	// Start building the query
	mods := []bob.Mod[*psql.SelectQuery]{
		sm.Columns("template_local_id", "template_name", "template_format", "template_id", "entity_id", "sender_id", "message_type"),
		sm.From("msg_template"),
	}

	// Add conditions using multiple Where clauses
	if msgtemplate.TemplateLocalID != 0 {
		mods = append(mods, sm.Where(psql.Quote("template_local_id").EQ(psql.Arg(msgtemplate.TemplateLocalID))))
	}
	if msgtemplate.ApplicationID != "" {
		mods = append(mods, sm.Where(psql.Quote("application_id").EQ(psql.Arg(msgtemplate.ApplicationID))))
	}
	if msgtemplate.TemplateFormat != "" {
		mods = append(mods, sm.Where(psql.Quote("template_format").EQ(psql.Arg(msgtemplate.TemplateFormat))))
	}

	query := psql.Select(mods...)

	// Execute the transaction
	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.GetTemplateformatbyID], &listTemplates)
		if err != nil {
			log.Error(ctx, "Error executing query in GetTemplateformatbyID repo function:  %s", err.Error())
			return err
		}
		return nil
	})

	if TxDB != nil {
		log.Error(ctx, "Error initiating transaction in GetTemplateformatbyID repo function:  %s", TxDB.Error())
		return nil, TxDB
	}

	return listTemplates, nil
}
*/
