package config

//Real Time Filtering by device/alertid/or other tags

// InfluxCfg is the main configuration for any InfluxDB TSDB
type InfluxCfg struct {
	ID                 string `xorm:"'id' unique" binding:"Required"`
	Host               string `xorm:"host" binding:"Required"`
	Port               int    `xorm:"port" binding:"Required;IntegerNotZero"`
	DB                 string `xorm:"db" binding:"Required"`
	User               string `xorm:"user" binding:"Required"`
	Password           string `xorm:"password" binding:"Required"`
	Retention          string `xorm:"'retention' default 'autogen'" binding:"Required"`
	Precision          string `xorm:"'precision' default 's'" binding:"Default(s);OmitEmpty;In(h,m,s,ms,u,ns)"` //posible values [h,m,s,ms,u,ns] default seconds for the nature of data
	Timeout            int    `xorm:"'timeout' default 30" binding:"Default(30);IntegerNotZero"`
	UserAgent          string `xorm:"useragent" binding:"Default(pseriescollector)"`
	EnableSSL          bool   `xorm:"enable_ssl"`
	SSLCA              string `xorm:"ssl_ca"`
	SSLCert            string `xorm:"ssl_cert"`
	SSLKey             string `xorm:"ssl_key"`
	InsecureSkipVerify bool   `xorm:"insecure_skip_verify"`
	Description        string `xorm:"description"`
}

//http://www-01.ibm.com/support/docview.wss?uid=nas8N1019111

// HMCCfg contains all hmcrelated device definitions
type HMCCfg struct {
	ID string `xorm:"'id' unique" binding:"Required"`
	//https://+Host+:12443
	Host     string `xorm:"host" binding:"Required"`
	Port     int    `xorm:"port" binding:"Required"`
	User     string `xorm:"user" binding:"Required"`
	Password string `xorm:"password" binding:"Required"`
	Samples  int    `xorm:"samples" binding:"Required"`

	Active bool `xorm:"'active' default 1"`

	Freq int `xorm:"'freq' default 60" binding:"Default(60);IntegerNotZero"`

	OutDB    string `xorm:"outdb"`
	LogLevel string `xorm:"loglevel" binding:"Default(info)"`
	LogFile  string `xorm:"logfile"`

	//influx tags
	DeviceTagName  string   `xorm:"devicetagname" binding:"Default(hostname)"`
	DeviceTagValue string   `xorm:"devicetagvalue" binding:"Default(id)"`
	ExtraTags      []string `xorm:"extra-tags"`

	Description string `xorm:"description"`

	//Filters for measurements
	//	MeasurementGroups []string `xorm:"-"`
	//	MeasFilters       []string `xorm:"-"`
}

// TableName go-xorm way to set the Table name to something different to "alert_h_t_t_p_out_rel"
func (HMCCfg) TableName() string {
	return "hmc_cfg"
}

// DBConfig read from DB
type DBConfig struct {
	Influxdb map[string]*InfluxCfg
	HMC      map[string]*HMCCfg
}

// Init initialices the DB
func Init(cfg *DBConfig) error {

	log.Debug("--------------------Initializing Config-------------------")

	log.Debug("-----------------------END Config metrics----------------------")
	return nil
}
