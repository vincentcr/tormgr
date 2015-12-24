: ${ENV:="dev"}
: ${APP_NAME:="tormgr"}
: ${DB_NAME:="${APP_NAME}_${ENV}"}
: ${DB_USER:="${APP_NAME}_${ENV}"}
: ${DB_PASSWD:="${APP_NAME}_${ENV}_secret"}

if [ -n "${DB_HOST}" ]; then
	HOST_ARG="-h ${DB_HOST}"
fi

SQL_DIR='./sql'

PSQL_CMD="psql -v ON_ERROR_STOP=1"

function main {
	case $1 in
    views)
    views
    ;;
    schema)
    schema
    ;;
    drop)
    drop
    ;;
    recreate)
    recreate
    ;;
    dev-data)
    dev_data
    ;;
    *)
    create
    ;;
	esac
}

function dev_data() {
	run_sql "${SQL_DIR}/dev.sql"
}

function views() {
	run_sql "${SQL_DIR}/views.sql"
}

function schema() {
	run_sql "${SQL_DIR}/schema.sql"
	views
}

function create() {
	run_su_sql " --set=dbuser=${DB_USER} --set=dbpasswd=${DB_PASSWD} --set=dbname=${DB_NAME} -f ${SQL_DIR}/db.sql"
	schema
	if [ "${ENV}" == "dev" ] ; then
		dev_data
	fi
}

function drop() {
	run_su_sql "-c 'DROP DATABASE ${DB_NAME}'"
	run_su_sql "-c 'DROP USER ${DB_USER}'"
}

function recreate() {
	drop
	create
}

function run_sql {
	PGPASSWORD=${DB_PASSWD} run_cmd "${PSQL_CMD} ${HOST_ARG} -d ${DB_NAME} -U ${DB_USER} -f $@"
}

function run_su_sql {
	PGPASSWORD=${POSTGRES_PASSWORD} run_cmd "${PSQL_CMD} ${HOST_ARG} -U postgres $@"
}

function run_cmd {
	echo "run: $@"
	eval "$@"
	res=$?
	if [ "${res}" != "0" ] ; then
		echo "command failed: $@ => ${res}"
		exit 1
	fi
}

main $1
