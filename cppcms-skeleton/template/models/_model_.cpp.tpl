/**
 * @PROJECT_NAME_HUMAN@  @DESCRIPTION@
 *
 * Copyright (C) @YEARS@ @AUTHOR@ <@EMAIL@>
 * See accompanying file COPYING.TXT file for licensing details.
 *
 * @category @PROJECT_NAME_HUMAN@
 * @author   @AUTHOR@ <@EMAIL@> 
 * @package  Models
 *
 */

#include <iostream>
#include <string>
#include <map>
#include <algorithm>
#include <limits>
#include <cppdb/frontend.h>

#include "models/%%MODEL_NAME%%.h"


namespace @PROJECT_NS@ {
namespace models {

/**
 *
 */
%%MODEL_NAME%%::%%MODEL_NAME%%() :
    MySQLModel()
{
}


} // end namespace models
} // end namespace @PROJECT_NS@


