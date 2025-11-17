package repository

import (
	"context"
	"errors"

	"MgApplication/core/domain"
	"MgApplication/core/port"

	config "MgApplication/api-config"
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

type ProviderRepository struct {
	Db  *dblib.DB
	Cfg *config.Config
}

// NewOfficeRepository preates a new Office repository instance
func NewProviderRepository(Db *dblib.DB, Cfg *config.Config) *ProviderRepository {
	return &ProviderRepository{
		Db,
		Cfg,
	}
}

func (pr *ProviderRepository) CreateMessageProviderRepo(gctx *gin.Context, msgprovider *domain.MsgProvider) (domain.MsgProvider, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), pr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	var msgProvider1 domain.MsgProvider
	TxDB := pr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		// Check if data already exists
		query1 := psql.Select(
			sm.Columns("COUNT(1) as count"),
			sm.From("msg_provider"),
			sm.Where(psql.Quote("provider_name").EQ(psql.Arg(msgprovider.ProviderName))),
		)
		err := dblib.TxReturnRow(ctx, tx, query1, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error checking whether a provider already exists with the inputted provider_name in CreateMsgProvider: %s", err.Error())
			return err
		}
		if Counter.Count > 0 {
			return errors.New("provider already exists, cannot save")
		}
		query2 := psql.Insert(
			im.Into("msg_provider", "provider_name", "short_name", "services", "configuration_key", "status_cd"),
			im.Values(psql.Arg(msgprovider.ProviderName, msgprovider.ShortName, msgprovider.Services, msgprovider.ConfigurationKeys, msgprovider.Status)),
			im.Returning("provider_id", "provider_name", "short_name", "services", "configuration_key", "status_cd"),
		)
		err = dblib.TxReturnRow(ctx, tx, query2, pgx.RowToStructByNameLax[domain.MsgProvider], &msgProvider1)
		if err != nil {
			log.Error(ctx, "Error executing insert query in CreateMsgProvider repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Transaction rolling back in CreateMsgProvider repo function:  %s", TxDB.Error())
		return domain.MsgProvider{}, TxDB
	}
	return msgProvider1, nil
}

/*
func (pr *ProviderRepository) ListProvidersTx(gctx *gin.Context) ([]domain.MsgProvider, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), pr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	var listProviders []domain.MsgProvider

	TxDB := pr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := psql.Select(
			sm.Columns("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd", "STRING_AGG(mr.request_type, ', ') AS services"),
			sm.From("msg_provider mp"),
			sm.Join("LATERAL unnest(string_to_array(mp.services, ',')) AS rt(rt_value) ON true"),
			sm.Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code"),
			sm.GroupBy("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd"),
			sm.OrderBy("mp.provider_id"),
		)

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.MsgProvider], &listProviders)
		if err != nil {
			log.Error(ctx, "Error executing query in ListProviders repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Error initiating transaction in ListProviders repo function:  %s", TxDB.Error())
		return nil, TxDB
	}
	return listProviders, nil
}

func (pr *ProviderRepository) ListProvidersOld(gctx *gin.Context) ([]domain.MsgProvider, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), pr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	query := psql.Select(
		sm.Columns("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd", "STRING_AGG(mr.request_type, ', ') AS services"),
		sm.From("msg_provider mp"),
		sm.Join("LATERAL unnest(string_to_array(mp.services, ',')) AS rt(rt_value) ON true"),
		sm.Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code"),
		sm.GroupBy("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd"),
		sm.OrderBy("mp.provider_id"),
	)
	return dblib.SelectRows(ctx, pr.Db, query, pgx.RowToStructByNameLax[domain.MsgProvider])
}
*/

func (pr *ProviderRepository) FetchMessageProviderRepo(gctx *gin.Context, msgprovider *domain.MsgProvider) ([]domain.MsgProvider, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), pr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var listProviders []domain.MsgProvider

	TxDB := pr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd", "STRING_AGG(mr.request_type, ', ') AS services").
			From("msg_provider mp").
			Join("LATERAL unnest(string_to_array(mp.services, ',')) AS rt(rt_value) ON true").
			Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code").
			Where(squirrel.Eq{"provider_id": msgprovider.ProviderID}).
			GroupBy("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd").
			OrderBy("mp.provider_id")

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.MsgProvider], &listProviders)
		if err != nil {
			log.Error(ctx, "Error executing query in GetProviderbyID repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Error initiating transaction in GetProviderbyID repo function:  %s", TxDB.Error())
		return nil, TxDB
	}
	return listProviders, nil
}

/*
func (pr *ProviderRepository) ListActiveProvidersTx(gctx *gin.Context) ([]domain.MsgProvider, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), pr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	var listProviders []domain.MsgProvider

	TxDB := pr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd", "STRING_AGG(mr.request_type, ', ') AS services").
			From("msg_provider mp").
			Join("LATERAL unnest(string_to_array(mp.services, ',')) AS rt(rt_value) ON true").
			Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code").
			Where(squirrel.Eq{"status_cd": 1}).
			GroupBy("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd").
			OrderBy("mp.provider_id")

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.MsgProvider], &listProviders)
		if err != nil {
			log.Error(ctx, "Error executing query in ListActiveProviders repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Error initiating transaction in ListActiveProviders repo function:  %s", TxDB.Error())
		return nil, TxDB
	}
	return listProviders, nil
}
*/

func (pr *ProviderRepository) UpdateMessageProviderRepo(gctx *gin.Context, msgprovider *domain.MsgProvider) (domain.MsgProvider, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), pr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	var msgProvider1 domain.MsgProvider
	TxDB := pr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		// Check if data already exists
		query1 := dblib.Psql.Select("COUNT(1) as count").
			From("msg_provider").
			Where(squirrel.Eq{"provider_id": msgprovider.ProviderID})
		err := dblib.TxReturnRow(ctx, tx, query1, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error checking whether a provider already exists with given provider_id in EditMsgProvider repo function:  %s", err.Error())
			return err
		}
		if Counter.Count == 0 {
			return errors.New("provider does not exists, cannot update")
		}
		query2 := dblib.Psql.Select("COUNT(1) as count").
			From("msg_provider").
			Where(squirrel.And{squirrel.Eq{"provider_name": msgprovider.ProviderName}, squirrel.NotEq{"provider_id": msgprovider.ProviderID}})
		err = dblib.TxReturnRow(ctx, tx, query2, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error checking whether a provider with the given provider name already exists for given provider_id in EditMsgProvider repo function:  %s", err.Error())
			return err
		}
		if Counter.Count > 0 {
			return errors.New("already one application with these selected details is available")
		}
		query3 := dblib.Psql.Update("msg_provider").
			Set("provider_name", msgprovider.ProviderName).
			Set("services", msgprovider.Services).
			Set("status_cd", msgprovider.Status).
			Where(squirrel.Eq{"provider_id": msgprovider.ProviderID}).
			Suffix("RETURNING provider_id,provider_name,services,status_cd")
		err = dblib.TxReturnRow(ctx, tx, query3, pgx.RowToStructByNameLax[domain.MsgProvider], &msgProvider1)
		if err != nil {
			log.Error(ctx, "Error executing Update query in EditMsgProvider repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Transaction rolling back in EditMsgProvider repo function:  %s", TxDB.Error())
		return domain.MsgProvider{}, TxDB
	}
	return msgProvider1, nil
}

func (pr *ProviderRepository) ToggleMessageProviderStatusRepo(gctx *gin.Context, msgprovider *domain.StatusProvider) (interface{}, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), pr.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	var CurrentStatus domain.CurrentStatus
	TxDB := pr.Db.WithTx(ctx, func(tx pgx.Tx) error {
		// Check if data already exists
		query1 := dblib.Psql.Select("COUNT(1) as count").
			From("msg_provider").
			Where(squirrel.Eq{"provider_id": msgprovider.ProviderID})
		err := dblib.TxReturnRow(ctx, tx, query1, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error checking whether provider already exists for the given provider_id in StatusMsgProvider repo function:  %s", err.Error())
			return err
		}
		if Counter.Count == 0 {
			return errors.New("no provider with selected details available")
		}
		// Retrieve the current status of the provider
		query2 := dblib.Psql.Select("status_cd").
			From("msg_provider").
			Where(squirrel.Eq{"provider_id": msgprovider.ProviderID})
		err = dblib.TxReturnRow(ctx, tx, query2, pgx.RowToStructByNameLax[domain.CurrentStatus], &CurrentStatus)
		if err != nil {
			log.Error(ctx, "Error checking the status of message provider in StatusMsgProvider repo function:  %s", err.Error())
			return err
		}
		if CurrentStatus.Status == 1 {
			// The current status is 1, check for dependent applications
			query3 := dblib.Psql.Select("COUNT(1) as count").
				From("msg_template").
				Where(squirrel.Eq{"gateway::int": msgprovider.ProviderID})
			err := dblib.TxReturnRow(ctx, tx, query3, pgx.RowToStructByNameLax[domain.Counter], &Counter)
			if err != nil {
				log.Error(ctx, "Error checking whether any template exists for the given provider in StatusMsgProvider repo function:  %s", err.Error())
				return err
			}
			if Counter.Count > 0 {
				// There are dependent applications with status 1, cannot update the provider status
				return errors.New("cannot disable the provider as there are some messages which are mapped to it")
			}
		}
		query4 := dblib.Psql.Update("msg_provider").
			Set("status_cd", squirrel.Expr("CASE WHEN status_cd = 0 THEN 1 ELSE 0 END")).
			Where(squirrel.Eq{"provider_id": msgprovider.ProviderID})
		err = dblib.TxExec(ctx, tx, query4)
		if err != nil {
			log.Error(ctx, "Error executing update query in StatusMsgProvider repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Transaction rolling back in StatusMsgProvider repo function:  %s", TxDB.Error())
		return map[string]interface{}{}, TxDB
	}
	return map[string]interface{}{}, nil
}

/*
func (pr *ProviderRepository) ListActiveProviders(gctx *gin.Context) ([]domain.MsgProvider, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), pr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	query := dblib.Psql.Select("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd", "STRING_AGG(mr.request_type, ', ') AS services").
		From("msg_provider mp").
		Join("LATERAL unnest(string_to_array(mp.services, ',')) AS rt(rt_value) ON true").
		Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code").
		Where(squirrel.Eq{"status_cd": 1}).
		GroupBy("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd").
		OrderBy("mp.provider_id")
	return dblib.SelectRows(ctx, pr.Db, query, pgx.RowToStructByNameLax[domain.MsgProvider])
}
*/

func (pr *ProviderRepository) ListMessageProvidersRepo(gctx *gin.Context, msgreq domain.ListMessageProviders, meta port.MetaDataRequest) ([]domain.MsgProvider, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), pr.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	// Build the base query
	query := dblib.Psql.Select("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd", "STRING_AGG(mr.request_type, ', ') AS services").
		From("msg_provider mp").
		Join("LATERAL unnest(string_to_array(mp.services, ',')) AS rt(rt_value) ON true").
		Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code")

		// Check the Status field for true, false, or nil
		// if msgreq.Status != nil {
	if msgreq.Status {
		query = query.Where(squirrel.Eq{"mp.status_cd": 1}) // Active applications
	}
	// else {
	// 	query = query.Where(squirrel.Eq{"mp.status_cd": 0}) // Inactive applications
	// }
	// }
	// Group and order the results
	query = query.GroupBy("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd").
		OrderBy("mp.provider_id").
		Offset(meta.Skip * meta.Limit).
		Limit(meta.Limit)

	// Execute the query and collect the rows
	collectedRows, err := dblib.SelectRows(ctx, pr.Db, query, pgx.RowToStructByNameLax[domain.MsgProvider])
	if err != nil {
		log.Error(ctx, "Error executing query in ListProvidersConditional repo function:  %s", err.Error())
		return nil, err
	}

	return collectedRows, nil
}
