#include <cppcms/json.h>
#include <booster/thread.h>
namespace cppcms {
namespace impl {
	struct cached_settings {
		struct cached_security {
			long long multipart_form_data_limit;
			long long content_length_limit;
			int file_in_memory_limit;
			std::string uploads_path;
			bool display_error_message;
			
			struct cached_csrf {
				bool enable;
				bool automatic;
				bool exposed;

				cached_csrf(json::value const &v)
				{
					enable = v.get("security.csrf.enable",false);
					automatic = v.get("security.csrf.automatic",true);
					exposed = v.get("security.csrf.exposed",false);
				}
			} csrf;

			cached_security(json::value const &v) :
				csrf(v)
			{
				multipart_form_data_limit = v.get("security.multipart_form_data_limit",64*1024);
				content_length_limit = v.get("security.content_length_limit",1024);
				file_in_memory_limit = v.get("security.file_in_memory_limit",128*1024);
				uploads_path = v.get("security.uploads_path","");
				display_error_message = v.get("security.display_error_message",false);
			}
		} security;
		struct cached_fastcgi {
			int cuncurrency_hint;
			cached_fastcgi(json::value const &v)
			{
				cuncurrency_hint = v.get("fastcgi.cuncurrency_hint",-1);
			}
		} fastcgi;
		struct cached_service {
			std::string ip;
			int port;
			int output_buffer_size;
			bool disable_xpowered_by;
			bool generate_http_headers;
			int worker_threads;
			int worker_processes;
			cached_service(json::value const &v)
			{
				ip = v.get("service.ip","127.0.0.1");
				port = v.get("service.port",8080);
				output_buffer_size = v.get("service.output_buffer_size",16384);
				disable_xpowered_by = v.get("service.disable_xpowered_by",false);
				unsigned cpus = booster::thread::hardware_concurrency();
				if(cpus == 0)
					cpus = 1;
				worker_threads = v.get("service.worker_threads",5 * cpus);
				worker_processes = v.get("service.worker_processes",0);
				generate_http_headers = v.get("service.generate_http_headers",false);
			}
		} service;
		struct cached_localization {
			bool disable_charset_in_content_type;
			cached_localization(json::value const &v) : 
				disable_charset_in_content_type(v.get("localization.disable_charset_in_content_type",false))
			{
			}
		} localization;
		struct cached_gzip {
			bool enable;
			int level;
			int buffer;
			cached_gzip(json::value const &v)
			{
				enable=v.get("gzip.enable",true);
				level=v.get("gzip.level",-1);
				buffer=v.get("gzip.buffer",-1);
			}
		} gzip;
		struct cached_http {
			struct cached_proxy {
				bool behind;
				std::vector<std::string> remote_addr_headers;
			} proxy;
			std::vector<std::string> script_names;
			int timeout;
			cached_http(json::value const &v) 
			{
				proxy.behind=v.get("http.proxy.behind",false);
				std::vector<std::string> default_headers;
				default_headers.push_back("X-Forwarded-For");
				proxy.remote_addr_headers = v.get("http.proxy.remote_addr_headers",default_headers);
				script_names = v.get("http.script_names",std::vector<std::string>());
				std::string script = v.get("http.script","");
				if(!script.empty())
					script_names.push_back(script);
				timeout = v.get("http.timeout",30);
			}
		} http;
		struct cached_session {
			int timeout;
			std::string expire;
			struct cached_cookies {
				std::string prefix;
				std::string domain;
				std::string path;
				bool secure;
			} cookies;
			cached_session(json::value const &v)
			{
				timeout = v.get("session.timeout",24*3600);
				expire = v.get("session.expire","browser");
				cookies.prefix = v.get("session.cookies.prefix","cppcms_session");
				cookies.domain = v.get("session.cookies.domain","");
				cookies.path = v.get("session.cookies.path","/");
				cookies.secure = v.get("session.cookies.secure",false);
			}
		} session;
		cached_settings(json::value const &v) :
			security(v),
			fastcgi(v),
			service(v),
			localization(v),
			gzip(v),
			http(v),
			session(v)	
		{
		}

	};
} // impl
} // cppcms

