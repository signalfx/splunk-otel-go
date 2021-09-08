// Package dbsystem provides identifiers for database systems conforming to
// OpenTelemetry semantic conventions.
package dbsystem

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Type is a database system type.
type Type attribute.KeyValue

// Attribute returns t as an attribute KeyValue. If t is empty the returned
// attribute will default to a Type OtherSQL.
func (t Type) Attribute() attribute.KeyValue {
	if t.Empty() {
		return semconv.DBSystemOtherSQL
	}
	return attribute.KeyValue(t)
}

// Empty returns if t is a defined database system Type of not.
func (t Type) Empty() bool {
	return !t.Key.Defined()
}

var (
	// Some other SQL database. Fallback only. See notes
	OtherSQL = Type(semconv.DBSystemOtherSQL)
	// Microsoft SQL Server
	MSSQL = Type(semconv.DBSystemMSSQL)
	// MySQL
	MySQL = Type(semconv.DBSystemMySQL)
	// Oracle Database
	Oracle = Type(semconv.DBSystemOracle)
	// IBM DB2
	DB2 = Type(semconv.DBSystemDB2)
	// PostgreSQL
	PostgreSQL = Type(semconv.DBSystemPostgreSQL)
	// Amazon Redshift
	Redshift = Type(semconv.DBSystemRedshift)
	// Apache Hive
	Hive = Type(semconv.DBSystemHive)
	// Cloudscape
	Cloudscape = Type(semconv.DBSystemCloudscape)
	// HyperSQL DataBase
	HSQLDB = Type(semconv.DBSystemHSQLDB)
	// Progress Database
	Progress = Type(semconv.DBSystemProgress)
	// SAP MaxDB
	MaxDB = Type(semconv.DBSystemMaxDB)
	// SAP HANA
	HanaDB = Type(semconv.DBSystemHanaDB)
	// Ingres
	Ingres = Type(semconv.DBSystemIngres)
	// FirstSQL
	FirstSQL = Type(semconv.DBSystemFirstSQL)
	// EnterpriseDB
	EDB = Type(semconv.DBSystemEDB)
	// InterSystems Caché
	Cache = Type(semconv.DBSystemCache)
	// Adabas (Adaptable Database System)
	Adabas = Type(semconv.DBSystemAdabas)
	// Firebird
	Firebird = Type(semconv.DBSystemFirebird)
	// Apache Derby
	Derby = Type(semconv.DBSystemDerby)
	// FileMaker
	Filemaker = Type(semconv.DBSystemFilemaker)
	// Informix
	Informix = Type(semconv.DBSystemInformix)
	// InstantDB
	InstantDB = Type(semconv.DBSystemInstantDB)
	// InterBase
	Interbase = Type(semconv.DBSystemInterbase)
	// MariaDB
	MariaDB = Type(semconv.DBSystemMariaDB)
	// Netezza
	Netezza = Type(semconv.DBSystemNetezza)
	// Pervasive PSQL
	Pervasive = Type(semconv.DBSystemPervasive)
	// PointBase
	Pointbase = Type(semconv.DBSystemPointbase)
	// SQLite
	Sqlite = Type(semconv.DBSystemSqlite)
	// Sybase
	Sybase = Type(semconv.DBSystemSybase)
	// Teradata
	Teradata = Type(semconv.DBSystemTeradata)
	// Vertica
	Vertica = Type(semconv.DBSystemVertica)
	// H2
	H2 = Type(semconv.DBSystemH2)
	// ColdFusion IMQ
	Coldfusion = Type(semconv.DBSystemColdfusion)
	// Apache Cassandra
	Cassandra = Type(semconv.DBSystemCassandra)
	// Apache HBase
	HBase = Type(semconv.DBSystemHBase)
	// MongoDB
	MongoDB = Type(semconv.DBSystemMongoDB)
	// Redis
	Redis = Type(semconv.DBSystemRedis)
	// Couchbase
	Couchbase = Type(semconv.DBSystemCouchbase)
	// CouchDB
	CouchDB = Type(semconv.DBSystemCouchDB)
	// Microsoft Azure Cosmos DB
	CosmosDB = Type(semconv.DBSystemCosmosDB)
	// Amazon DynamoDB
	DynamoDB = Type(semconv.DBSystemDynamoDB)
	// Neo4j
	Neo4j = Type(semconv.DBSystemNeo4j)
	// Apache Geode
	Geode = Type(semconv.DBSystemGeode)
	// Elasticsearch
	Elasticsearch = Type(semconv.DBSystemElasticsearch)
	// Memcached
	Memcached = Type(semconv.DBSystemMemcached)
	// CockroachDB
	Cockroachdb = Type(semconv.DBSystemCockroachdb)
)
