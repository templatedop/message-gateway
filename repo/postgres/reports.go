package repository

import (
	"context"
	"time"

	"MgApplication/core/domain"
	"MgApplication/core/port"

	config "MgApplication/api-config"
	dblib "MgApplication/api-db"
	log "MgApplication/api-log"

	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type ReportsRepository struct {
	Db  *dblib.DB
	Cfg *config.Config
}

// NewOfficeRepository creates a new Office repository instance
func NewReportsRepository(Db *dblib.DB, Cfg *config.Config) *ReportsRepository {
	return &ReportsRepository{
		Db,
		Cfg,
	}
}

func (cr *ReportsRepository) SMSSentStatusReportRepo(gctx *gin.Context, fromDate time.Time, toDate time.Time, meta port.MetaDataRequest) ([]domain.SMSReport, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), cr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	var sms []domain.SMSReport
	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("row_number() over(ORDER BY created_date ASC) as serial_number", "created_date", "communication_id", "application_id", "facility_id", "priority", "message_text", "unnest(mr.mobile_number) AS mobile_number", "gateway", "status").
			From("msg_request mr").
			Where(squirrel.And{squirrel.GtOrEq{"created_date::date": fromDate}, squirrel.LtOrEq{"created_date::date": toDate}}).
			OrderBy("created_date ASC").
			Offset(meta.Skip * meta.Limit).
			Limit(meta.Limit)

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.SMSReport], &sms)
		if err != nil {
			log.Error(gctx, "Error in SMSSentStatusReport repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in SMSSentStatusReport repo function:  %s", TxDB.Error())
		return nil, TxDB
	}
	return sms, nil
}

func (cr *ReportsRepository) AppwiseSMSUsageReportRepo(gctx *gin.Context, fromDate time.Time, toDate time.Time, meta port.MetaDataRequest) ([]domain.SMSAggregateReport, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), cr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	var sms []domain.SMSAggregateReport
	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("row_number() over(ORDER BY mr.created_date::date ASC) as serial_number",
			"ma.application_name",
			"mr.created_date::date",
			"COUNT(*) AS total_sms, COUNT(CASE WHEN mr.status = 'submitted' THEN 1 END) AS success, COUNT(CASE WHEN mr.status <> 'submitted' THEN 1 END) AS failed").
			From("msg_request mr").
			Join("msg_application ma ON mr.application_id::int = ma.application_id").
			Join("unnest(mr.mobile_number) AS mobile_number ON true").
			Where(squirrel.And{squirrel.GtOrEq{"mr.created_date::date": fromDate}, squirrel.LtOrEq{"mr.created_date::date": toDate}}).
			GroupBy("ma.application_name,mr.created_date::date").
			OrderBy("mr.created_date::date ASC").
			Offset(meta.Skip * meta.Limit).
			Limit(meta.Limit)

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.SMSAggregateReport], &sms)
		if err != nil {
			log.Error(gctx, "Error in Appwise SMS Usage Report repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in Appwise SMS Usage Report repo function:  %s", TxDB.Error())
		return nil, TxDB
	}

	return sms, nil
}

func (cr *ReportsRepository) TemplatewiseSMSUsageReportRepo(gctx *gin.Context, fromDate time.Time, toDate time.Time, meta port.MetaDataRequest) ([]domain.SMSAggregateReport, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), cr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	var sms []domain.SMSAggregateReport
	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("row_number() over(ORDER BY mr.created_date::date ASC) as serial_number", "ma.template_name", "mr.created_date::date", "COUNT(*) AS total_sms, COUNT(CASE WHEN mr.status = 'submitted' THEN 1 END) AS success, COUNT(CASE WHEN mr.status <> 'submitted' THEN 1 END) AS failed").
			From("msg_request mr").
			Join("msg_template ma ON mr.template_id = ma.template_id").
			Join("unnest(mr.mobile_number) AS mobile_number ON true").
			Where(squirrel.And{squirrel.GtOrEq{"mr.created_date::date": fromDate}, squirrel.LtOrEq{"mr.created_date::date": toDate}}).
			GroupBy("ma.template_name,mr.created_date::date").
			OrderBy("mr.created_date::date ASC").
			Offset(meta.Skip * meta.Limit).
			Limit(meta.Limit)

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.SMSAggregateReport], &sms)
		if err != nil {
			log.Error(gctx, "Error in Templatewise SMS Usage Report repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in Templatewise SMS Usage Report repo function:  %s", TxDB.Error())
		return nil, TxDB
	}

	return sms, nil
}

func (cr *ReportsRepository) ProviderwiseSMSUsageReportRepo(gctx *gin.Context, fromDate time.Time, toDate time.Time, meta port.MetaDataRequest) ([]domain.SMSAggregateReport, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), cr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	var sms []domain.SMSAggregateReport
	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("row_number() over(ORDER BY mr.created_date::date ASC) as serial_number", "ma.provider_name", "mr.created_date::date", "COUNT(*) AS total_sms, COUNT(CASE WHEN mr.status = 'submitted' THEN 1 END) AS success, COUNT(CASE WHEN mr.status <> 'submitted' THEN 1 END) AS failed").
			From("msg_request mr").
			Join("msg_provider ma ON mr.gateway::int = ma.provider_id").
			Join("unnest(mr.mobile_number) AS mobile_number ON true").
			Where(squirrel.And{squirrel.GtOrEq{"mr.created_date::date": fromDate}, squirrel.LtOrEq{"mr.created_date::date": toDate}}).
			GroupBy("ma.provider_name,mr.created_date::date").
			OrderBy("mr.created_date::date ASC").
			Offset(meta.Skip * meta.Limit).
			Limit(meta.Limit)

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.SMSAggregateReport], &sms)
		if err != nil {
			log.Error(gctx, "Error in Provider wise SMS Usage Report repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in Provider wise SMS Usage Report repo function:  %s", TxDB.Error())
		return nil, TxDB
	}

	return sms, nil
}

/*
// with transaction model
func (cr *ReportsRepository) SMSDashboard2(gctx *gin.Context) (domain.SMSDashboard, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), cr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	var sms domain.SMSDashboard

	TxDB := cr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select(
			//"COUNT(*) as total_requests",
			//"count(*) as total_sms_sent",
			"SUM(COALESCE(array_length(mr.mobile_number, 1), 0)) as total_sms_sent",
			//"sum(case when mr.priority=1 then 1 else 0 end) as total_otps",
			"SUM(CASE WHEN mr.priority=1 THEN COALESCE(array_length(mr.mobile_number, 1), 0) ELSE 0 END) as total_otps",
			//"sum(case when mr.priority=2 then 1 else 0 end) as total_transactions",
			"SUM(CASE WHEN mr.priority=2 THEN COALESCE(array_length(mr.mobile_number, 1), 0) ELSE 0 END) as total_transactions",
			//"sum(case when mr.priority=3 then 1 else 0 end) as total_bulk_sms",
			"SUM(CASE WHEN mr.priority=3 THEN COALESCE(array_length(mr.mobile_number, 1), 0) ELSE 0 END) as total_bulk_sms",
			//"sum(case when mr.priority=4 then 1 else 0 end) as total_promotional_sms",
			"SUM(CASE WHEN mr.priority=4 THEN COALESCE(array_length(mr.mobile_number, 1), 0) ELSE 0 END) as total_promotional_sms",
			"(select Count(*) from msg_template as mt where mt.status_cd=1)as total_templates",
			"(select count(*) from msg_provider mp where mp.status_cd=1) as total_providers",
			"(select count(*) from msg_application ma where ma.status_cd=1) as total_applications").
			From("msg_request as mr")

		err := dblib.TxReturnRow(ctx, tx, query, pgx.RowToStructByNameLax[domain.SMSDashboard], &sms)
		if err != nil {
			log.Error(gctx, "Error in SMS Dashboard repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(gctx, "Error initiating transaction in SMSDashboard repo function:  %s", TxDB.Error())
		return domain.SMSDashboard{}, TxDB
	}
	return sms, nil
}
*/

// without transaction model
func (cr *ReportsRepository) SMSDashboardRepo(gctx *gin.Context) (domain.SMSDashboard, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), cr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	query := dblib.Psql.Select(
		//"COUNT(*) as total_requests",
		//"count(*) as total_sms_sent",
		"SUM(COALESCE(array_length(mr.mobile_number, 1), 0)) as total_sms_sent",
		//"sum(case when mr.priority=1 then 1 else 0 end) as total_otps",
		"SUM(CASE WHEN mr.priority=1 THEN COALESCE(array_length(mr.mobile_number, 1), 0) ELSE 0 END) as total_otps",
		//"sum(case when mr.priority=2 then 1 else 0 end) as total_transactions",
		"SUM(CASE WHEN mr.priority=2 THEN COALESCE(array_length(mr.mobile_number, 1), 0) ELSE 0 END) as total_transactions",
		//"sum(case when mr.priority=3 then 1 else 0 end) as total_bulk_sms",
		"SUM(CASE WHEN mr.priority=3 THEN COALESCE(array_length(mr.mobile_number, 1), 0) ELSE 0 END) as total_bulk_sms",
		//"sum(case when mr.priority=4 then 1 else 0 end) as total_promotional_sms",
		"SUM(CASE WHEN mr.priority=4 THEN COALESCE(array_length(mr.mobile_number, 1), 0) ELSE 0 END) as total_promotional_sms",
		"(select Count(*) from msg_template as mt where mt.status_cd=1)as total_templates",
		"(select count(*) from msg_provider mp where mp.status_cd=1) as total_providers",
		"(select count(*) from msg_application ma where ma.status_cd=1) as total_applications").
		From("msg_request as mr")
	return dblib.SelectOne(ctx, cr.Db, query, pgx.RowToStructByNameLax[domain.SMSDashboard])
}
