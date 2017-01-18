/**
 * Copyright (C) 2016 Jimes Yang (sndnvaps) <admin@sndnvaps.com>
 * See accompanying file COPYING.TXT file for licensing details.
 *
 * @category Cppcms-skeleton
 * @author   Jimes Yang <admin@sndnvaps.com>
 * @package  Models
 *
 */

#include <fstream>

#include <booster/log.h>
#include <mysql/mysql.h>
#include "cppcms_skel/generics/Config.h"

#include "MySQLModel.h"
#include "SqlImporter.h"

namespace cppcmsskel {
namespace models {


/**
 *
 */
MySQLModel::MySQLModel() {
	create_session(
        Config::get_instance()->mysql_host,
	Config::get_instance()->mysql_user,
	Config::get_instance()->mysql_password,
	Config::get_instance()->mysql_database
    );
}


/**
 *
 */
MySQLModel::MySQLModel(cppdb::session MySQLDbParam) : mysqlDb(MySQLDbParam) {
}

MySQLModel::MySQLModel(	const std::string &host,
			const std::string &user,
			const std::string &password,
			const std::string &database) {
	create_session(host,user,password,database);
}

void MySQLModel::create_session(const std::string &host,
				const std::string &user,
				const std::string &password,
				const std::string &database) {
	try {
		mysqlDb = cppdb::session(
			"mysql:host=" + host + ";" + "database=" + database + ";"
			"user=" + user + ";" + "password=" + password + ";"
			"@pool_size=16");
	    } catch (cppdb::cppdb_error const &e) {
			BOOSTER_ERROR("cppcms") << e.what();
	}

}

/**
 *
 */
bool MySQLModel::import_sql_file(
    const std::string &sqlFilePath
) {

    SqlImporter importer(mysqlDb);
    return importer.from_file(sqlFilePath);
}

/**
 *
 */
bool MySQLModel::execute_simple(
    cppdb::statement &statement
) {
    try {
        statement.exec();
    } catch (cppdb::cppdb_error const& e) {
        BOOSTER_ERROR("cppcms") << e.what();
        statement.reset();
        return false;
    }
    statement.reset();
    return true;
}

/**
 *
 */
bool MySQLModel::check_existence(
    cppdb::statement &statement
) {
    cppdb::result res = statement.row();
    // Don't forget to reset statement
    statement.reset();
   if (!res.empty()) {
	return true;
     }
    return false;
}


} // end of namespace models
} // end namespace cppcmsskel

