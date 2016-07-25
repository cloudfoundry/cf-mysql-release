package logwriter

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

type LogWriter interface {
	Write(string) error
}

type logWriter struct {
	db      *sql.DB
	logPath string
}

func New(db *sql.DB, logPath string) LogWriter {
	return &logWriter{
		db:      db,
		logPath: logPath,
	}
}

func (lw *logWriter) Write(ts string) error {
	var statusColumnNames []string
	var statusColumnValues []string

	statusQuery := `SHOW STATUS WHERE Variable_name IN (
		'wsrep_ready',
		'wsrep_cluster_conf_id',
		'wsrep_cluster_status',
		'wsrep_connected',
		'wsrep_local_state_comment',
		'wsrep_local_recv_queue_avg',
		'wsrep_flow_control_paused',
		'wsrep_cert_deps_distance',
		'wsrep_local_send_queue_avg',
		'wsrep_last_committed'
		)`
	status, err := lw.db.Query(statusQuery)

	if err != nil {
		return err
	}

	defer status.Close()
	for status.Next() {
		var varName string
		var varValue string
		status.Scan(&varName, &varValue)
		statusColumnNames = append(statusColumnNames, varName)
		statusColumnValues = append(statusColumnValues, varValue)
	}
	info, err := os.Stat(lw.logPath)
	writeHeaders := false
	if err != nil || info.Size() == 0 {
		writeHeaders = true
	}

	f, err := os.OpenFile(lw.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	columnNamesStr := strings.Join(statusColumnNames, ",")
	columnValuesStr := strings.Join(statusColumnValues, ",")

	if writeHeaders {
		f.WriteString(fmt.Sprintf("%s,%s", "timestamp", columnNamesStr))
		f.WriteString("\n")
	}
	f.WriteString(fmt.Sprintf("%s,%s", ts, columnValuesStr))
	f.WriteString("\n")
	return nil
}
