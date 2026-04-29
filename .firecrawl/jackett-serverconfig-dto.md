using System;
using System.Collections.Generic;
using System.Runtime.Serialization;
using Jackett.Common.Models.Config;

namespace Jackett.Common.Models.DTO
{
 \[DataContract\]
 public class ServerConfig
 {
 \[DataMember\]
 public IEnumerable notices { get; set; }
 \[DataMember\]
 public int port { get; set; }
 \[DataMember\]
 public bool external { get; set; }
 \[DataMember\]
 public string local\_bind\_address { get; set; }
 \[DataMember\]
 public bool cors { get; set; }
 \[DataMember\]
 public string api\_key { get; set; }
 \[DataMember\]
 public string blackholedir { get; set; }
 \[DataMember\]
 public bool updatedisabled { get; set; }
 \[DataMember\]
 public bool prerelease { get; set; }
 \[DataMember\]
 public string password { get; set; }
 \[DataMember\]
 public bool logging { get; set; }
 \[DataMember\]
 public string basepathoverride { get; set; }
 \[DataMember\]
 public string baseurloverride { get; set; }
 \[DataMember\]
 public bool cache\_enabled { get; set; }
 \[DataMember\]
 public long cache\_ttl { get; set; }
 \[DataMember\]
 public long cache\_max\_results\_per\_indexer { get; set; }
 \[DataMember\]
 public string flaresolverrurl { get; set; }
 \[DataMember\]
 public int flaresolverr\_maxtimeout { get; set; }
 \[DataMember\]
 public string omdbkey { get; set; }
 \[DataMember\]
 public string omdburl { get; set; }
 \[DataMember\]
 public string app\_version { get; set; }
 \[DataMember\]
 public bool can\_run\_netcore { get; set; }

 \[DataMember\]
 public ProxyType proxy\_type { get; set; }
 \[DataMember\]
 public string proxy\_url { get; set; }
 \[DataMember\]
 public int? proxy\_port { get; set; }
 \[DataMember\]
 public string proxy\_username { get; set; }
 \[DataMember\]
 public string proxy\_password { get; set; }

 public ServerConfig()
 {
 notices = Array.Empty();
 }

 public ServerConfig(IEnumerable notices, Models.Config.ServerConfig config, string version, bool canRunNetCore)
 {
 this.notices = notices;
 port = config.Port;
 external = config.AllowExternal;
 local\_bind\_address = config.LocalBindAddress;
 cors = config.AllowCORS;
 api\_key = config.APIKey;
 blackholedir = config.BlackholeDir;
 updatedisabled = config.UpdateDisabled;
 prerelease = config.UpdatePrerelease;
 password = string.IsNullOrEmpty(config.AdminPassword) ? string.Empty : config.AdminPassword.Substring(0, 10);
 logging = config.RuntimeSettings.TracingEnabled;
 basepathoverride = config.BasePathOverride;
 baseurloverride = config.BaseUrlOverride;
 cache\_enabled = config.CacheEnabled;
 cache\_ttl = config.CacheTtl;
 cache\_max\_results\_per\_indexer = config.CacheMaxResultsPerIndexer;
 flaresolverrurl = config.FlareSolverrUrl;
 flaresolverr\_maxtimeout = config.FlareSolverrMaxTimeout;
 omdbkey = config.OmdbApiKey;
 omdburl = config.OmdbApiUrl;
 app\_version = version;
 can\_run\_netcore = canRunNetCore;

 proxy\_type = config.ProxyType;
 proxy\_url = config.ProxyUrl;
 proxy\_port = config.ProxyPort;
 proxy\_username = config.ProxyUsername;
 proxy\_password = config.ProxyPassword;
 }
 }
}