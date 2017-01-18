/**
 * Copyright (C) 2016 Jimes Yang (sndnvaps) <admin@sndnvaps.com>
 * See accompanying file COPYING.TXT file for licensing details.
 *
 * @category Cppcms-skeleton
 * @author   Jimes Yang <admin@sndnvaps.com>
 * @package  Models
 *
 */

#ifndef CPPCMS_SKEL_MODELS_MYSQL_H
#define CPPCMS_SKEL_MODELS_MYSQL_H

#include <cppdb/frontend.h>

namespace cppcmsskel {
namespace models {

/**
 * @class MySQLModel
 * 
 * @brief base class to represent a model based on
 *        a MySQL database 
 */

class MySQLModel {
	private:
		void create_session(
			const std::string &host,
			const std::string &user,
			const std::string &password,
			const std::string &database
		);
	protected:
		cppdb::session mysqlDb;
		//TODO
		bool execute_simple(
			cppdb::statement &statement
		);

        /**
         * @brief Used with a statement try to check the existence
         *        or not of a given record
         *
         * @param statement The statement to execute
         *
         * @return bool True if the record exists, false otherwise
         *
         * @since 25 April 2012
         */
		bool check_existence(
			cppdb::statement &statement
		);

	public:
		MySQLModel();
	        /**
         * @brief Create a Model based on sqlite3 database
         *
         * @param databasePath The location of the Sqlite3 database
         *
         * @since 3 January 2013
         */
		MySQLModel(
			const std::string &host,
			const std::string &user,
			const std::string &password,
			const std::string &database
		);
		MySQLModel(cppdb::session mysqldb);


        /**
         * @brief Import an SQL file into the database opened by the model
         *        FIXME TODO for the moment the import will totally ignore what's
         *        written after the last ';' 
         *
         * @param sqlFilePath The location of the file to import 
         *
         * @return bool True if the file is correctly imported, false otherwise
         */
		bool import_sql_file(
			const std::string &sqlFilePath
		);
};

}
}
#endif
