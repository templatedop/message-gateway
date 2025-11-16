package repository

import (
	"context"
	"errors"
	"fmt"

	"MgApplication/core/domain"

	config "MgApplication/api-config"
	dblib "MgApplication/api-db"
	log "MgApplication/api-log"

	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type TemplateRepository struct {
	Db  *dblib.DB
	Cfg *config.Config
}

func NewTemplateRepository(Db *dblib.DB, Cfg *config.Config) *TemplateRepository {
	return &TemplateRepository{
		Db,
		Cfg,
	}
}

func (tr *TemplateRepository) CreateTemplateRepo(gctx *gin.Context, mtemplate *domain.MaintainTemplate) error {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	TxDB := tr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		// Check if data already exists
		query := dblib.Psql.Select("COUNT(1) as count").
			From("msg_template").
			Where(squirrel.Eq{"template_id": mtemplate.TemplateID})
		err := dblib.TxReturnRow(ctx, tx, query, pgx.RowToStructByPos[domain.Counter], &Counter)

		if err != nil {
			log.Error(gctx, "Error checking whether a msg template exists for the given template_id in MaintainTemplate repo function:  %s", err.Error())
			return err
		}
		if Counter.Count > 0 {
			return errors.New("given template_id and template already exists, cannot continue")
		}
		uquery := dblib.Psql.Insert("msg_template").
			Columns("application_id", "template_name", "template_format", "entity_id", "sender_id", "template_id", "gateway", "message_type", "status_cd").
			Values(mtemplate.ApplicationID, mtemplate.TemplateName, mtemplate.TemplateFormat, mtemplate.EntityID, mtemplate.SenderID, mtemplate.TemplateID, mtemplate.Gateway, mtemplate.MessageType, mtemplate.Status)
		err = dblib.TxExec(ctx, tx, uquery)
		if err != nil {
			log.Error(gctx, "Error executing insert query in MaintainTemplate repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Transaction rolling back in MaintainTemplate repo function:  %s", TxDB.Error())
		return TxDB
	}
	return nil
}

/*
func (tr *TemplateRepository) ListTemplates2(gctx *gin.Context) ([]domain.MaintainTemplate, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	var listTemplates []domain.MaintainTemplate

	TxDB := tr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("mt.template_local_id", "STRING_AGG(ma.application_name, ', ') AS application_id", "mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id", "mt.template_id", "mp.provider_name AS gateway", "mt.status_cd").
			From("msg_template mt").
			Join("LATERAL unnest(string_to_array(mt.application_id, ',')) AS rt(rt_value) ON true").
			Join("msg_application ma ON rt.rt_value::integer = ma.application_id").
			Join("msg_provider mp on mp.provider_id=mt.gateway::integer").
			GroupBy("mt.template_local_id", "mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id", "mt.template_id", "mp.provider_name", "mt.status_cd").
			OrderBy("mt.template_local_id")

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.MaintainTemplate], &listTemplates)
		if err != nil {
			log.Error(gctx, "Error executing query in ListTemplates repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in ListTemplates repo function:  %s", TxDB.Error())
		return nil, TxDB
	}
	return listTemplates, nil
}

func (tr *TemplateRepository) ListTemplatesOld(gctx *gin.Context) ([]domain.MaintainTemplate, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	query := dblib.Psql.Select("mt.template_local_id", "STRING_AGG(ma.application_name, ', ') AS application_id", "mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id", "mt.template_id", "mt.message_type", "mp.provider_name AS gateway", "mt.status_cd").
		From("msg_template mt").
		Join("LATERAL unnest(string_to_array(mt.application_id, ',')) AS rt(rt_value) ON true").
		Join("msg_application ma ON rt.rt_value::integer = ma.application_id").
		Join("msg_provider mp on mp.provider_id=mt.gateway::integer").
		GroupBy("mt.template_local_id", "mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id", "mt.template_id", "mt.message_type", "mp.provider_name", "mt.status_cd").
		OrderBy("mt.template_local_id")
	return dblib.SelectRows(ctx, tr.Db, query, pgx.RowToStructByNameLax[domain.MaintainTemplate])
}

func (tr *TemplateRepository) ListTemplatesLimit(gctx *gin.Context, listTemplate *domain.Meta) ([]domain.MaintainTemplate, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	query := dblib.Psql.Select("mt.template_local_id", "STRING_AGG(ma.application_name, ', ') AS application_id",
		"mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id", "mt.template_id", "mt.message_type",
		"mp.provider_name AS gateway", "mt.status_cd").
		From("msg_template mt").
		Join("LATERAL unnest(string_to_array(mt.application_id, ',')) AS rt(rt_value) ON true").
		Join("msg_application ma ON rt.rt_value::integer = ma.application_id").
		Join("msg_provider mp on mp.provider_id=mt.gateway::integer").
		GroupBy("mt.template_local_id", "mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id", "mt.template_id", "mt.message_type", "mp.provider_name", "mt.status_cd").
		OrderBy("mt.template_local_id").
		Limit(listTemplate.Limit).
		Offset(listTemplate.Skip)

	return dblib.SelectRows(ctx, tr.Db, query, pgx.RowToStructByNameLax[domain.MaintainTemplate])
}
*/

func (tr *TemplateRepository) ListTemplatesRepo(gctx *gin.Context, listTemplate *domain.Meta) ([]domain.MaintainTemplate, uint64, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	var totalCount uint64

	// Create the subquery for counting total templates
	subquery, _, _ := dblib.Psql.Select("COUNT(*) AS total_count").
		From("msg_template").
		ToSql()

	// Build the main query to fetch the templates with pagination and total_count from the subquery
	query := dblib.Psql.Select("mt.template_local_id", "STRING_AGG(ma.application_name, ', ') AS application_id",
		"mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id", "mt.template_id",
		"mt.message_type", "mp.provider_name AS gateway", "mt.status_cd", fmt.Sprintf("(%s) AS total_count", subquery)).
		From("msg_template mt").
		Join("LATERAL unnest(string_to_array(mt.application_id, ',')) AS rt(rt_value) ON true").
		Join("msg_application ma ON rt.rt_value::integer = ma.application_id").
		Join("msg_provider mp on mp.provider_id=mt.gateway::integer").
		GroupBy("mt.template_local_id", "mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id",
			"mt.template_id", "mt.message_type", "mp.provider_name", "mt.status_cd").
		OrderBy("mt.template_local_id").
		Limit(uint64(listTemplate.Limit)).
		Offset(uint64(listTemplate.Skip))

	// Execute the main query to fetch templates and total count
	templates, err := dblib.SelectRows(ctx, tr.Db, query, pgx.RowToStructByNameLax[domain.MaintainTemplate])
	if err != nil {
		log.Error(gctx, "DB Error in ListTemplatesLimit: %s", err.Error())
		return nil, 0, err
	}

	// Fetch the total count using the subquery
	if len(templates) > 0 {
		totalCount = templates[0].TotalCount
	}

	// Return the templates and total count
	return templates, totalCount, nil
}

func (tr *TemplateRepository) ToggleTemplateStatusRepo(gctx *gin.Context, msgtemplate *domain.StatusTemplate) (interface{}, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	TxDB := tr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		// Check if data already exists
		query := dblib.Psql.Select("COUNT(1) as count").
			From("msg_template").
			Where(squirrel.Eq{"template_local_id": msgtemplate.TemplateLocalID})
		err := dblib.TxReturnRow(ctx, tx, query, pgx.RowToStructByPos[domain.Counter], &Counter)

		if err != nil {
			log.Error(gctx, "Error checking whether a msg_template exists for the given template_local_id in StatusTemplate repo function: %s", err.Error())
			return err
		}
		if Counter.Count == 0 {
			return errors.New("no template with selected details is available")
		}
		uquery := dblib.Psql.Update("msg_template").
			Set("status_cd", squirrel.Expr("CASE WHEN status_cd = 0 THEN 1 ELSE 0 END")).
			Where(squirrel.Eq{"template_local_id": msgtemplate.TemplateLocalID})
		err = dblib.TxExec(ctx, tx, uquery)
		if err != nil {
			log.Error(gctx, "Error executing update query in StatusTemplate repo function: %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Transaction rolling back in Status Template repo function:  %s", TxDB.Error())
		return map[string]interface{}{}, TxDB
	}
	return map[string]interface{}{}, nil
}

/*
func (tr *TemplateRepository) GetTemplatebyID2(gctx *gin.Context, msgtemplate *domain.MaintainTemplate) ([]domain.MaintainTemplate, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var listTemplates []domain.MaintainTemplate

	TxDB := tr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("mt.template_local_id", "STRING_AGG(ma.application_name, ', ') AS application_id", "mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id", "mt.template_id", "mt.gateway", "mt.status_cd").
			From("msg_template mt").
			Join("LATERAL unnest(string_to_array(mt.application_id, ',')) AS rt(rt_value) ON true").
			Join("msg_application ma ON rt.rt_value::integer = ma.application_id").
			Join("msg_provider mp on mp.provider_id=mt.gateway::integer").
			Where(squirrel.Eq{"template_local_id": msgtemplate.TemplateLocalID}).
			GroupBy("mt.template_local_id", "mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id", "mt.template_id", "mp.provider_name", "mt.status_cd").
			OrderBy("mt.template_local_id")

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.MaintainTemplate], &listTemplates)
		if err != nil {
			log.Error(gctx, "Error executing query in GetTemplatebyID repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in GetTemplatebyID repo function:  %s", TxDB.Error())
		return nil, TxDB
	}
	return listTemplates, nil
}
*/

func (tr *TemplateRepository) FetchTemplateRepo(gctx *gin.Context, msgtemplate *domain.MaintainTemplate) ([]domain.MaintainTemplate, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	query := dblib.Psql.Select("mt.template_local_id", "STRING_AGG(ma.application_name, ', ') AS application_id", "mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id", "mt.template_id", "mt.message_type", "mt.gateway", "mt.status_cd").
		From("msg_template mt").
		Join("LATERAL unnest(string_to_array(mt.application_id, ',')) AS rt(rt_value) ON true").
		Join("msg_application ma ON rt.rt_value::integer = ma.application_id").
		Join("msg_provider mp on mp.provider_id=mt.gateway::integer").
		Where(squirrel.Eq{"template_local_id": msgtemplate.TemplateLocalID}).
		GroupBy("mt.template_local_id", "mt.template_name", "mt.template_format", "mt.sender_id", "mt.entity_id", "mt.template_id", "mt.message_type", "mp.provider_name", "mt.status_cd").
		OrderBy("mt.template_local_id")
	return dblib.SelectRows(ctx, tr.Db, query, pgx.RowToStructByNameLax[domain.MaintainTemplate])
}

func (tr *TemplateRepository) UpdateTemplateRepo(gctx *gin.Context, msgtemplate *domain.MaintainTemplate) error {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	TxDB := tr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		// Check if data already exists
		query := dblib.Psql.Select("COUNT(1) as count").
			From("msg_template").
			Where(squirrel.Eq{"template_local_id": msgtemplate.TemplateLocalID})
		err := dblib.TxReturnRow(ctx, tx, query, pgx.RowToStructByPos[domain.Counter], &Counter)

		if err != nil {
			log.Error(gctx, "Error checking whether a msg_template exists for the given template_local_id in EditTemplate repo function: %s", err.Error())
			return err
		}
		if Counter.Count == 0 {
			return errors.New("template does not exists, cannot update")
		}
		uquery := dblib.Psql.Update("msg_template").
			Set("application_id", msgtemplate.ApplicationID).
			Set("template_name", msgtemplate.TemplateName).
			Set("template_format", msgtemplate.TemplateFormat).
			Set("sender_id", msgtemplate.SenderID).
			Set("entity_id", msgtemplate.EntityID).
			Set("template_id", msgtemplate.TemplateID).
			Set("gateway", msgtemplate.Gateway).
			Set("message_type", msgtemplate.MessageType).
			Set("status_cd", msgtemplate.Status).
			Where(squirrel.Eq{"template_local_id": msgtemplate.TemplateLocalID})
		err = dblib.TxExec(ctx, tx, uquery)
		if err != nil {
			log.Error(gctx, "Error executing update query in EditTemplate repo function: %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Transaction rolling back in EditTemplate repo function:  %s", TxDB.Error())
		return TxDB
	}
	return nil
}

func (tr *TemplateRepository) FetchTemplateByApplicationRepo(gctx *gin.Context, msgtemplate *domain.MaintainTemplate) ([]domain.GetTemplatebyAPPID, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var listTemplates []domain.GetTemplatebyAPPID

	TxDB := tr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("mt.template_local_id", "mt.template_name").
			From("msg_template mt").
			Join("LATERAL unnest(string_to_array(mt.application_id, ',')) AS rt(rt_value) ON true").
			Join("msg_application ma ON rt.rt_value::integer = ma.application_id").
			Join("msg_provider mp on mp.provider_id=mt.gateway::integer").
			Where("mt.status_cd = 1").                                                               // Add condition for status=1
			Where("ARRAY["+msgtemplate.ApplicationID+"]::integer[] @> ARRAY[rt.rt_value::integer]"). // Check if given application_id exists in the array
			GroupBy("mt.template_local_id", "mt.template_name").
			OrderBy("mt.template_local_id")

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.GetTemplatebyAPPID], &listTemplates)
		if err != nil {
			log.Error(gctx, "Error executing query in GetTemplatenamesbyID repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in GetTemplatenamesbyID repo function:  %s", TxDB.Error())
		return nil, TxDB
	}
	return listTemplates, nil
}

/*
func (tr *TemplateRepository) GetTemplateformatbyID(gctx *gin.Context, msgtemplate *domain.MaintainTemplate) ([]domain.GetTemplateformatbyID, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var listTemplates []domain.GetTemplateformatbyID

	TxDB := tr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("template_local_id", "template_name", "template_format", "template_id", "entity_id", "sender_id", "message_type").
			From("msg_template").
			Where(squirrel.Eq{"template_local_id": msgtemplate.TemplateLocalID})

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.GetTemplateformatbyID], &listTemplates)
		if err != nil {
			log.Error(gctx, "Error executing query in GetTemplateformatbyID repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in GetTemplateformatbyID repo function:  %s", TxDB.Error())
		return nil, TxDB
	}
	return listTemplates, nil
}


func (tr *TemplateRepository) GetTemplateIDbyformatRepo(gctx *gin.Context, msgtemplate *domain.MaintainTemplate) ([]domain.GetTemplateIDbyformat, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var listTemplates []domain.GetTemplateIDbyformat

	TxDB := tr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("template_id").
			From("msg_template").
			Where(squirrel.Eq{"template_format": msgtemplate.TemplateFormat})

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.GetTemplateIDbyformat], &listTemplates)
		if err != nil {
			log.Error(gctx, "Error executing query in GetTemplateIDbyformat repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in GetTemplateIDbyformat repo function:  %s", TxDB.Error())
		return nil, TxDB
	}
	return listTemplates, nil
}

func (tr *TemplateRepository) GetSenderIDbyTemplateformat(gctx *gin.Context, msgtemplate *domain.MaintainTemplate) ([]domain.GetSenderIDbyTemplateformat, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var listTemplates []domain.GetSenderIDbyTemplateformat

	TxDB := tr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("sender_id").
			From("msg_template").
			Where(squirrel.Eq{"template_format": msgtemplate.TemplateFormat})

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.GetSenderIDbyTemplateformat], &listTemplates)
		if err != nil {
			log.Error(gctx, "Error executing query in GetSenderIDbyTemplateformat repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in GetSenderIDbyTemplateformat repo function:  %s", TxDB.Error())
		return nil, TxDB
	}
	return listTemplates, nil
}
*/

func (tr *TemplateRepository) FetchTemplateDetailsRepo(gctx *gin.Context, msgtemplate *domain.MaintainTemplate) ([]domain.GetTemplateformatbyID, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), tr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var listTemplates []domain.GetTemplateformatbyID

	// Start building the query
	query := dblib.Psql.Select("template_local_id", "template_name", "template_format", "template_id", "entity_id", "sender_id", "message_type").
		From("msg_template")

	// Add conditions using multiple Where clauses
	if msgtemplate.TemplateLocalID != 0 {
		query = query.Where(squirrel.Eq{"template_local_id": msgtemplate.TemplateLocalID})
	}
	if msgtemplate.ApplicationID != "" {
		query = query.Where(squirrel.Eq{"application_id": msgtemplate.ApplicationID})
	}
	if msgtemplate.TemplateFormat != "" {
		query = query.Where(squirrel.Eq{"template_format": msgtemplate.TemplateFormat})
	}

	// Execute the transaction
	TxDB := tr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.GetTemplateformatbyID], &listTemplates)
		if err != nil {
			log.Error(gctx, "Error executing query in GetTemplateformatbyID repo function: %s", err.Error())
			return err
		}
		return nil
	})

	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in GetTemplateformatbyID repo function: %s", TxDB.Error())
		return nil, TxDB
	}

	return listTemplates, nil
}
