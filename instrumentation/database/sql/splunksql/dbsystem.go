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

package splunksql

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// DBSystem is a database system type.
type DBSystem attribute.KeyValue

// Attribute returns db as an attribute KeyValue. If db is empty the returned
// attribute will default to a Type OtherSQL.
func (db DBSystem) Attribute() attribute.KeyValue {
	if db.Empty() {
		return semconv.DBSystemOtherSQL
	}
	return attribute.KeyValue(db)
}

// Empty returns if db is a defined database system or not.
func (db DBSystem) Empty() bool {
	return !db.Key.Defined()
}

var (
	// DBSystemOtherSQL is some other SQL database. This is used as a fallback
	// only.
	DBSystemOtherSQL = DBSystem(semconv.DBSystemOtherSQL)
	// DBSystemMSSQL is a Microsoft SQL Server database system.
	DBSystemMSSQL = DBSystem(semconv.DBSystemMSSQL)
	// DBSystemMySQL is a MySQL database system.
	DBSystemMySQL = DBSystem(semconv.DBSystemMySQL)
	// DBSystemOracle is an Oracle Database database system.
	DBSystemOracle = DBSystem(semconv.DBSystemOracle)
	// DBSystemDB2 is a IBM DB2 database system.
	DBSystemDB2 = DBSystem(semconv.DBSystemDB2)
	// DBSystemPostgreSQL is a PostgreSQL database system.
	DBSystemPostgreSQL = DBSystem(semconv.DBSystemPostgreSQL)
	// DBSystemRedshift is an Amazon Redshift database system.
	DBSystemRedshift = DBSystem(semconv.DBSystemRedshift)
	// DBSystemHive is an Apache Hive database system.
	DBSystemHive = DBSystem(semconv.DBSystemHive)
	// DBSystemCloudscape is a Cloudscape database system.
	DBSystemCloudscape = DBSystem(semconv.DBSystemCloudscape)
	// DBSystemHSQLDB is a HyperSQL DataBase database system.
	DBSystemHSQLDB = DBSystem(semconv.DBSystemHSQLDB)
	// DBSystemProgress is a Progress Database database system.
	DBSystemProgress = DBSystem(semconv.DBSystemProgress)
	// DBSystemMaxDB is an SAP MaxDB database system.
	DBSystemMaxDB = DBSystem(semconv.DBSystemMaxDB)
	// DBSystemHanaDB is an SAP HANA database system.
	DBSystemHanaDB = DBSystem(semconv.DBSystemHanaDB)
	// DBSystemIngres is an Ingres database system.
	DBSystemIngres = DBSystem(semconv.DBSystemIngres)
	// DBSystemFirstSQL is a FirstSQL database system.
	DBSystemFirstSQL = DBSystem(semconv.DBSystemFirstSQL)
	// DBSystemEDB is an EnterpriseDB database system.
	DBSystemEDB = DBSystem(semconv.DBSystemEDB)
	// DBSystemCache is an InterSystems Caché database system.
	DBSystemCache = DBSystem(semconv.DBSystemCache)
	// DBSystemAdabas is an Adabas (Adaptable Database System) database
	// system.
	DBSystemAdabas = DBSystem(semconv.DBSystemAdabas)
	// DBSystemFirebird is a Firebird database system.
	DBSystemFirebird = DBSystem(semconv.DBSystemFirebird)
	// DBSystemDerby is an Apache Derby database system.
	DBSystemDerby = DBSystem(semconv.DBSystemDerby)
	// DBSystemFilemaker is a FileMaker database system.
	DBSystemFilemaker = DBSystem(semconv.DBSystemFilemaker)
	// DBSystemInformix is an Informix database system.
	DBSystemInformix = DBSystem(semconv.DBSystemInformix)
	// DBSystemInstantDB is an InstantDB database system.
	DBSystemInstantDB = DBSystem(semconv.DBSystemInstantDB)
	// DBSystemInterbase is an InterBase database system.
	DBSystemInterbase = DBSystem(semconv.DBSystemInterbase)
	// DBSystemMariaDB is a MariaDB database system.
	DBSystemMariaDB = DBSystem(semconv.DBSystemMariaDB)
	// DBSystemNetezza is a Netezza database system.
	DBSystemNetezza = DBSystem(semconv.DBSystemNetezza)
	// DBSystemPervasive is a Pervasive PSQL database system.
	DBSystemPervasive = DBSystem(semconv.DBSystemPervasive)
	// DBSystemPointbase is a PointBase database system.
	DBSystemPointbase = DBSystem(semconv.DBSystemPointbase)
	// DBSystemSqlite is a SQLite database system.
	DBSystemSqlite = DBSystem(semconv.DBSystemSqlite)
	// DBSystemSybase is a Sybase database system.
	DBSystemSybase = DBSystem(semconv.DBSystemSybase)
	// DBSystemTeradata is a Teradata database system.
	DBSystemTeradata = DBSystem(semconv.DBSystemTeradata)
	// DBSystemVertica is a Vertica database system.
	DBSystemVertica = DBSystem(semconv.DBSystemVertica)
	// DBSystemH2 is a H2 database system.
	DBSystemH2 = DBSystem(semconv.DBSystemH2)
	// DBSystemColdfusion is a ColdFusion IMQ database system.
	DBSystemColdfusion = DBSystem(semconv.DBSystemColdfusion)
	// DBSystemCassandra is an Apache Cassandra database system.
	DBSystemCassandra = DBSystem(semconv.DBSystemCassandra)
	// DBSystemHBase is an Apache HBase database system.
	DBSystemHBase = DBSystem(semconv.DBSystemHBase)
	// DBSystemMongoDB is a MongoDB database system.
	DBSystemMongoDB = DBSystem(semconv.DBSystemMongoDB)
	// DBSystemRedis is a Redis database system.
	DBSystemRedis = DBSystem(semconv.DBSystemRedis)
	// DBSystemCouchbase is a Couchbase database system.
	DBSystemCouchbase = DBSystem(semconv.DBSystemCouchbase)
	// DBSystemCouchDB is a CouchDB database system.
	DBSystemCouchDB = DBSystem(semconv.DBSystemCouchDB)
	// DBSystemCosmosDB is a Microsoft Azure Cosmos DB database system.
	DBSystemCosmosDB = DBSystem(semconv.DBSystemCosmosDB)
	// DBSystemDynamoDB is an Amazon DynamoDB database system.
	DBSystemDynamoDB = DBSystem(semconv.DBSystemDynamoDB)
	// DBSystemNeo4j is a Neo4j database system.
	DBSystemNeo4j = DBSystem(semconv.DBSystemNeo4j)
	// DBSystemGeode is an Apache Geode database system.
	DBSystemGeode = DBSystem(semconv.DBSystemGeode)
	// DBSystemElasticsearch is an Elasticsearch database system.
	DBSystemElasticsearch = DBSystem(semconv.DBSystemElasticsearch)
	// DBSystemMemcached is a Memcached database system.
	DBSystemMemcached = DBSystem(semconv.DBSystemMemcached)
	// DBSystemCockroachdb is a CockroachDB database system.
	DBSystemCockroachdb = DBSystem(semconv.DBSystemCockroachdb)
)
