# ! /usr/bin/python2.7
# apt-get install -y python-mysqldb
import MySQLdb
from subprocess import Popen,PIPE
sqlfile = "config/db.sql"
sqlfile1 = "config/init.sql"
host = "localhost"
usr = "root"
passwd = "you_password_here"
database = "studygolang"
port = 3306

try:
	conn = MySQLdb.connect(host=host,user=usr,passwd=passwd,port=port)
        cur = conn.cursor()
	result = cur.execute('CREATE DATABASE IF NOT EXISTS studygolang;')
        result = cur.execute('SELECT roleid, name, op_user,ctime,mtime FROM studygolang.role limit 0;')
	print result
        cur.close()
        conn.close()
except MySQLdb.Error,e:
	print "Mysql Error %d: %s" % (e.args[0], e.args[1])
	process = Popen('mysql -h%s -P%s -u%s -p%s %s'  %(host, port, usr, passwd, database), stdout=PIPE, stdin=PIPE, shell=True)
	output = process.communicate('source '+sqlfile)
	process = Popen('mysql -h%s -P%s -u%s -p%s %s'  %(host, port, usr, passwd, database), stdout=PIPE, stdin=PIPE, shell=True)
	output = process.communicate('source '+sqlfile1)
	print output


