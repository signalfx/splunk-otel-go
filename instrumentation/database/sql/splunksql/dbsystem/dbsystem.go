// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package dbsystem provides identifiers for database systems conforming to
// OpenTelemetry semantic conventions.
package dbsystem // import "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/dbsystem"

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
	// OtherSQL is some other SQL database. This is used as a fallback only.
	OtherSQL = Type(semconv.DBSystemOtherSQL)
	// MSSQL is a Microsoft SQL Server database system.
	MSSQL = Type(semconv.DBSystemMSSQL)
	// MySQL is a MySQL database system.
	MySQL = Type(semconv.DBSystemMySQL)
	// Oracle is an Oracle Database database system.
	Oracle = Type(semconv.DBSystemOracle)
	// DB2 is a IBM DB2 database system.
	DB2 = Type(semconv.DBSystemDB2)
	// PostgreSQL is a PostgreSQL database system.
	PostgreSQL = Type(semconv.DBSystemPostgreSQL)
	// Redshift is an Amazon Redshift database system.
	Redshift = Type(semconv.DBSystemRedshift)
	// Hive is an Apache Hive database system.
	Hive = Type(semconv.DBSystemHive)
	// Cloudscape is a Cloudscape database system.
	Cloudscape = Type(semconv.DBSystemCloudscape)
	// HSQLDB is a HyperSQL DataBase database system.
	HSQLDB = Type(semconv.DBSystemHSQLDB)
	// Progress is a Progress Database database system.
	Progress = Type(semconv.DBSystemProgress)
	// MaxDB is an SAP MaxDB database system.
	MaxDB = Type(semconv.DBSystemMaxDB)
	// HanaDB is an SAP HANA database system.
	HanaDB = Type(semconv.DBSystemHanaDB)
	// Ingres is an Ingres database system.
	Ingres = Type(semconv.DBSystemIngres)
	// FirstSQL is a FirstSQL database system.
	FirstSQL = Type(semconv.DBSystemFirstSQL)
	// EDB is an EnterpriseDB database system.
	EDB = Type(semconv.DBSystemEDB)
	// Cache is an InterSystems Caché database system.
	Cache = Type(semconv.DBSystemCache)
	// Adabas is an Adabas (Adaptable Database System) database system.
	Adabas = Type(semconv.DBSystemAdabas)
	// Firebird is a Firebird database system.
	Firebird = Type(semconv.DBSystemFirebird)
	// Derby is an Apache Derby database system.
	Derby = Type(semconv.DBSystemDerby)
	// Filemaker is a FileMaker database system.
	Filemaker = Type(semconv.DBSystemFilemaker)
	// Informix is an Informix database system.
	Informix = Type(semconv.DBSystemInformix)
	// InstantDB is an InstantDB database system.
	InstantDB = Type(semconv.DBSystemInstantDB)
	// Interbase is an InterBase database system.
	Interbase = Type(semconv.DBSystemInterbase)
	// MariaDB is a MariaDB database system.
	MariaDB = Type(semconv.DBSystemMariaDB)
	// Netezza is a Netezza database system.
	Netezza = Type(semconv.DBSystemNetezza)
	// Pervasive is a Pervasive PSQL database system.
	Pervasive = Type(semconv.DBSystemPervasive)
	// Pointbase is a PointBase database system.
	Pointbase = Type(semconv.DBSystemPointbase)
	// Sqlite is a SQLite database system.
	Sqlite = Type(semconv.DBSystemSqlite)
	// Sybase is a Sybase database system.
	Sybase = Type(semconv.DBSystemSybase)
	// Teradata is a Teradata database system.
	Teradata = Type(semconv.DBSystemTeradata)
	// Vertica is a Vertica database system.
	Vertica = Type(semconv.DBSystemVertica)
	// H2 is a H2 database system.
	H2 = Type(semconv.DBSystemH2)
	// Coldfusion is a ColdFusion IMQ database system.
	Coldfusion = Type(semconv.DBSystemColdfusion)
	// Cassandra is an Apache Cassandra database system.
	Cassandra = Type(semconv.DBSystemCassandra)
	// HBase is an Apache HBase database system.
	HBase = Type(semconv.DBSystemHBase)
	// MongoDB is a MongoDB database system.
	MongoDB = Type(semconv.DBSystemMongoDB)
	// Redis is a Redis database system.
	Redis = Type(semconv.DBSystemRedis)
	// Couchbase is a Couchbase database system.
	Couchbase = Type(semconv.DBSystemCouchbase)
	// CouchDB is a CouchDB database system.
	CouchDB = Type(semconv.DBSystemCouchDB)
	// CosmosDB is a Microsoft Azure Cosmos DB database system.
	CosmosDB = Type(semconv.DBSystemCosmosDB)
	// DynamoDB is an Amazon DynamoDB database system.
	DynamoDB = Type(semconv.DBSystemDynamoDB)
	// Neo4j is a Neo4j database system.
	Neo4j = Type(semconv.DBSystemNeo4j)
	// Geode is an Apache Geode database system.
	Geode = Type(semconv.DBSystemGeode)
	// Elasticsearch is an Elasticsearch database system.
	Elasticsearch = Type(semconv.DBSystemElasticsearch)
	// Memcached is a Memcached database system.
	Memcached = Type(semconv.DBSystemMemcached)
	// Cockroachdb is a CockroachDB database system.
	Cockroachdb = Type(semconv.DBSystemCockroachdb)
)
