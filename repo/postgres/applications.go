package repository

import (
	"context"
	"errors"

	"MgApplication/core/domain"
	"MgApplication/core/port"

	config "MgApplication/api-config"
	dblib "MgApplication/api-db"
	log "MgApplication/api-log"

	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type ApplicationRepository struct {
	Db  *dblib.DB
	Cfg *config.Config
}

// NewOfficeRepository creates a new Office repository instance
func NewApplicationRepository(Db *dblib.DB, Cfg *config.Config) *ApplicationRepository {
	return &ApplicationRepository{
		Db,
		Cfg,
	}
}

// Create MsgApplication a new MsgApplication record in the database
func (ar *ApplicationRepository) CreateMsgApplicationRepo(ctx context.Context, msgapp *domain.MsgApplications) (domain.MsgApplications, error) {

	ctx, cancel := context.WithTimeout(ctx, ar.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	var msgapplication domain.MsgApplications
	TxDB := ar.Db.WithTx(ctx, func(tx pgx.Tx) error {
		// Check if data already exists
		query1 := dblib.Psql.Select("COUNT(1) as count").
			From("msg_application").
			Where(squirrel.Eq{"application_name": msgapp.ApplicationName})
		err := dblib.TxReturnRow(ctx, tx, query1, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error checking whether an application exists or not in CreateMsgApplication repo function: %s", err.Error())
			return err
		}
		if Counter.Count > 0 {
			return errors.New("data already exists for this application")
		}
		query2 := dblib.Psql.Insert("msg_application").
			Columns("application_name", "request_type", "secret_key", "status_cd").
			Values(msgapp.ApplicationName, msgapp.RequestType, msgapp.SecretKey, msgapp.Status).
			Suffix("RETURNING application_id,application_name,request_type,created_date,updated_date,status_cd")
		err = dblib.TxReturnRow(ctx, tx, query2, pgx.RowToStructByNameLax[domain.MsgApplications], &msgapplication)
		if err != nil {
			log.Error(ctx, "Error executing insert query in CreateMsgApplication repo function: %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Transaction rolling back in StatusMsgApplication repo function: %s", TxDB.Error())
		return domain.MsgApplications{}, TxDB
	}
	return msgapplication, nil
}

/*
func (ar *ApplicationRepository) ListApplicationsTx(gctx *gin.Context) ([]domain.MsgApplicationsGet, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), ar.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	var listApplications []domain.MsgApplicationsGet

	TxDB := ar.Db.WithTx(ctx, func(tx pgx.Tx) error {
		query := dblib.Psql.Select("ma.application_id", "ma.application_name", "ma.status_cd", "STRING_AGG(mr.request_type, ', ') AS request_type").
			From("msg_application ma").
			Join("LATERAL unnest(string_to_array(ma.request_type, ',')) AS rt(rt_value) ON true").
			Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code").
			GroupBy("ma.application_id", "ma.application_name", "ma.status_cd").
			OrderBy("ma.application_id")

		err := dblib.TxRows(ctx, tx, query, pgx.RowToStructByNameLax[domain.MsgApplicationsGet], &listApplications)
		if err != nil {
			log.Error(ctx, "Error executing query in ListApplications repo function: %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Error initiating transaction in ListApplications repo function: %s", TxDB.Error())
		return nil, TxDB
	}
	return listApplications, nil
}

func (ar *ApplicationRepository) ListApplicationsOld(gctx *gin.Context) ([]domain.MsgApplicationsGet, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), ar.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	query := dblib.Psql.Select("ma.application_id", "ma.application_name", "ma.status_cd", "STRING_AGG(mr.request_type, ', ') AS request_type").
		From("msg_application ma").
		Join("LATERAL unnest(string_to_array(ma.request_type, ',')) AS rt(rt_value) ON true").
		Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code").
		GroupBy("ma.application_id", "ma.application_name", "ma.status_cd").
		OrderBy("ma.application_id")
	return dblib.SelectRows(ctx, ar.Db, query, pgx.RowToStructByNameLax[domain.MsgApplicationsGet])
}
*/

func (ar *ApplicationRepository) FetchApplicationRepo(ctx context.Context, msgapp *domain.MsgApplications) ([]domain.MsgApplicationsGet, error) {

	ctx, cancel := context.WithTimeout(ctx, ar.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var listApplications []domain.MsgApplicationsGet

	// TxDB := ar.Db.WithTx(ctx, func(tx pgx.Tx) error {
	query := dblib.Psql.Select("ma.application_id", "ma.application_name", "ma.status_cd", "STRING_AGG(mr.request_type, ', ') AS request_type").
		From("msg_application ma").
		Join("LATERAL unnest(string_to_array(ma.request_type, ',')) AS rt(rt_value) ON true").
		Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code").
		Where(squirrel.Eq{"application_id": msgapp.ApplicationID}).
		GroupBy("ma.application_id", "ma.application_name", "ma.status_cd").
		OrderBy("ma.application_id")

	listApplications, err := dblib.SelectRows(ctx, ar.Db, query, pgx.RowToStructByNameLax[domain.MsgApplicationsGet])
	if err != nil {
		log.Error(ctx, "Error executing query in GetAppbyID repo function:  %s", err.Error())
		return nil, err
	}
	// return nil
	// })
	// if TxDB != nil {
	// 	log.Error(ctx, "Error initiating transaction in GetAppbyID repo function: %s", TxDB.Error())
	// 	return nil, TxDB
	// }
	return listApplications, nil
}

/*
func (ar *ApplicationRepository) ListActiveApplications(gctx *gin.Context) ([]domain.MsgApplicationsGet, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), ar.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()
	query := dblib.Psql.Select("ma.application_id", "ma.application_name", "ma.status_cd", "STRING_AGG(mr.request_type, ', ') AS request_type").
		From("msg_application ma").
		Join("LATERAL unnest(string_to_array(ma.request_type, ',')) AS rt(rt_value) ON true").
		Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code").
		GroupBy("ma.application_id", "ma.application_name", "ma.status_cd").
		Where(squirrel.Eq{"status_cd": 1}).
		OrderBy("ma.application_id")
	return dblib.SelectRows(ctx, ar.Db, query, pgx.RowToStructByNameLax[domain.MsgApplicationsGet])
}

func (ar *ApplicationRepository) FetchApplications(gctx *gin.Context, applicationID uint64, activeOnly bool) ([]domain.MsgApplicationsGet, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), ar.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	// Build the base query
	query := dblib.Psql.Select("ma.application_id", "ma.application_name", "ma.status_cd", "STRING_AGG(mr.request_type, ', ') AS request_type").
		From("msg_application ma").
		Join("LATERAL unnest(string_to_array(ma.request_type, ',')) AS rt(rt_value) ON true").
		Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code")

	// Apply conditions based on the parameters
	if applicationID != 0 {
		// If an application ID is provided, filter by it
		query = query.Where(squirrel.Eq{"ma.application_id": applicationID})
	} else if activeOnly {
		// If only active applications are requested
		query = query.Where(squirrel.Eq{"ma.status_cd": 1})
	}

	query = query.GroupBy("ma.application_id", "ma.application_name", "ma.status_cd").
		OrderBy("ma.application_id")

	// Execute the query and return the results using dblib.SelectRows
	collectedRows, err := dblib.SelectRows(ctx, ar.Db, query, pgx.RowToStructByNameLax[domain.MsgApplicationsGet])
	if err != nil {
		log.Error(ctx, "Error executing query in FetchApplications repo function:  %s", err.Error())
		return nil, err
	}

	return collectedRows, nil
}
*/

func (ar *ApplicationRepository) UpdateMsgApplicationRepo(ctx context.Context, msgapp *domain.EditApplication) (domain.EditApplication, error) {

	ctx, cancel := context.WithTimeout(ctx, ar.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	var msgapplication domain.EditApplication
	TxDB := ar.Db.WithTx(ctx, func(tx pgx.Tx) error {
		// Check if data already exists
		query1 := dblib.Psql.Select("COUNT(1) as count").
			From("msg_application").
			Where(squirrel.Eq{"application_id": msgapp.ApplicationID})
		err := dblib.TxReturnRow(ctx, tx, query1, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error checking whether an application already exists or not in EditMsgApplication repo function:  %s", err.Error())
			return err
		}
		if Counter.Count == 0 {
			log.Error(ctx, "No application with selected details are available")
			return errors.New("no application with selected details available")
		}
		query2 := dblib.Psql.Select("COUNT(1) as count").
			From("msg_application").
			Where(squirrel.And{squirrel.Eq{"application_name": msgapp.ApplicationName}, squirrel.NotEq{"application_id": msgapp.ApplicationID}})
		err = dblib.TxReturnRow(ctx, tx, query2, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error executing select query in EditMsgApplication repo function:  %s", err.Error())
			return err
		}
		if Counter.Count > 0 {
			log.Error(ctx, "Already One application with the selected details already exists")
			return errors.New("already one application with these selected details is available")
		}
		query3 := dblib.Psql.Update("msg_application").
			Set("application_name", msgapp.ApplicationName).
			Set("request_type", msgapp.RequestType).
			Set("status_cd", msgapp.Status).
			Set("updated_date", squirrel.Expr("current_timestamp")).
			Where(squirrel.Eq{"application_id": msgapp.ApplicationID}).
			Suffix("RETURNING application_id,application_name,request_type,updated_date,status_cd")
		err = dblib.TxReturnRow(ctx, tx, query3, pgx.RowToStructByNameLax[domain.EditApplication], &msgapplication)
		if err != nil {
			log.Error(ctx, "Error executing update query in EditMsgApplication repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Transaction rolling back in EditMsgApplication repo function:  %s", TxDB.Error())
		return domain.EditApplication{}, TxDB
	}
	return msgapplication, nil
}

func (ar *ApplicationRepository) ToggleApplicationStatusRepo(gctx *gin.Context, msgapp *domain.StatusApplication) (interface{}, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), ar.Cfg.GetDuration("db.querytimeoutlow"))
	defer cancel()

	var Counter domain.Counter
	TxDB := ar.Db.WithTx(ctx, func(tx pgx.Tx) error {
		// Check if data already exists
		query1 := dblib.Psql.Select("COUNT(1) as count").
			From("msg_application").
			Where(squirrel.Eq{"application_id": msgapp.ApplicationID})
		err := dblib.TxReturnRow(ctx, tx, query1, pgx.RowToStructByNameLax[domain.Counter], &Counter)
		if err != nil {
			log.Error(ctx, "Error checking whether an application exists or not in StatusMsgApplication repo function:  %s", err.Error())
			return err
		}
		if Counter.Count == 0 {
			return errors.New("no application with selected details available")
		}
		query2 := dblib.Psql.Update("msg_application").
			Set("status_cd", squirrel.Expr("CASE WHEN status_cd = 0 THEN 1 ELSE 0 END")).
			Set("updated_date", squirrel.Expr("current_timestamp")).
			Where(squirrel.Eq{"application_id": msgapp.ApplicationID})
		err = dblib.TxExec(ctx, tx, query2)
		if err != nil {
			log.Error(ctx, "Error executing update query in StatusMsgApplication repo function:  %s", err.Error())
			return err
		}
		return nil
	})
	if TxDB != nil {
		log.Error(ctx, "Transaction rolling back in StatusMsgApplication repo function:  %s", TxDB.Error())
		return map[string]interface{}{}, TxDB
	}
	return map[string]interface{}{}, nil
}

/*
func (ar *ApplicationRepository) ListActiveProvidersTx(gctx *gin.Context) ([]domain.MsgProvider, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), ar.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	var listProviders []domain.MsgProvider

	TxDB := ar.Db.WithTx(ctx, func(tx pgx.Tx) error {
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

func (ar *ApplicationRepository) ListActiveProviders(gctx *gin.Context) ([]domain.MsgProvider, error) {

	ctx, cancel := context.WithTimeout(gctx.Request.Context(), ar.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	query := dblib.Psql.Select("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd", "STRING_AGG(mr.request_type, ', ') AS services").
		From("msg_provider mp").
		Join("LATERAL unnest(string_to_array(mp.services, ',')) AS rt(rt_value) ON true").
		Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code").
		Where(squirrel.Eq{"status_cd": 1}).
		GroupBy("mp.provider_id", "mp.provider_name", "mp.short_name", "mp.status_cd").
		OrderBy("mp.provider_id")
	return dblib.SelectRows(ctx, ar.Db, query, pgx.RowToStructByNameLax[domain.MsgProvider])
}
*/

func (ar *ApplicationRepository) ListApplicationsRepo(ctx context.Context, msgapp domain.ListApplications, meta port.MetaDataRequest) ([]domain.MsgApplicationsGet, error) {

	ctx, cancel := context.WithTimeout(ctx, ar.Cfg.GetDuration("db.querytimeoutmed"))
	defer cancel()

	// Build the base query
	query := dblib.Psql.Select("ma.application_id", "ma.application_name", "ma.status_cd", "STRING_AGG(mr.request_type, ', ') AS request_type").
		From("msg_application ma").
		Join("LATERAL unnest(string_to_array(ma.request_type, ',')) AS rt(rt_value) ON true").
		Join("msg_request_type mr ON rt.rt_value::integer = mr.request_code")

		// Check the Status field for true, false, or nil
		// if msgapp.Status != nil {
		// 	if *msgapp.Status {
		// 		query = query.Where(squirrel.Eq{"ma.status_cd": 1}) // Active applications
		// 	} else {
		// 		query = query.Where(squirrel.Eq{"ma.status_cd": 0}) // Inactive applications
		// 	}
		// }

		// Check the Status field for true, false
		//if msgapp.Status != nil {
	if msgapp.Status {
		query = query.Where(squirrel.Eq{"ma.status_cd": 1}) // Active applications
	}
	// else {
	// 	query = query.Where(squirrel.Eq{"ma.status_cd": 0}) // Inactive applications
	// }
	//}

	query = query.GroupBy("ma.application_id", "ma.application_name", "ma.status_cd").
		OrderBy("ma.application_id").
		Offset(meta.Skip * meta.Limit).
		Limit(meta.Limit)

	// Convert query to SQL string and log it
	sql, args, err := query.ToSql()
	if err != nil {
		log.Error(ctx, "Error generating SQL query: %s", err.Error())
		return nil, err
	}
	log.Debug(ctx, "SQL Query in ListApplicationsRepo: %s, Args: %v", sql, args)

	// Execute the query and collect the rows
	collectedRows, err := dblib.SelectRows(ctx, ar.Db, query, pgx.RowToStructByNameLax[domain.MsgApplicationsGet])
	if err != nil {
		log.Error(ctx, "Error executing query in GetApplications repo function:  %s", err.Error())
		return nil, err
	}

	return collectedRows, nil
}
