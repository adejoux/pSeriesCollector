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

	Active             bool `xorm:"'active' default 1"`
	ManagedSystemsOnly bool `xorm:"'managed_systems_only' default 0"`

	Freq           int `xorm:"'freq' default 60" binding:"Default(60);IntegerNotZero"`
	UpdateScanFreq int `xorm:"'update_scan_freq' default 60" binding:"Default(60);UIntegerAndLessOne"`

	OutDB       string `xorm:"outdb"`
	LogLevel    string `xorm:"loglevel" binding:"Default(info)"`
	HMCAPIDebug bool   `xorm:"hmc_api_debug"`
	LogFile     string `xorm:"logfile"`

	//influx tags
	DeviceTagName  string   `xorm:"devicetagname" binding:"Default(hostname)"`
	DeviceTagValue string   `xorm:"devicetagvalue" binding:"Default(id)"`
	ExtraTags      []string `xorm:"extra-tags"`

	Description string `xorm:"description"`
}

// TableName go-xorm way to set the Table name to something different to "alert_h_t_t_p_out_rel"
func (HMCCfg) TableName() string {
	return "hmc_cfg"
}

// DEVICE TABLE

// DeviceCfg contains all hmc related device definitions
type DeviceCfg struct {
	//LogicalPartition
	ID             string `xorm:"'id' unique" binding:"Required"` //lpar.PartitionUUID
	Name           string `xorm:"name" binding:"Required"`        //lpar.PartitionName
	SerialNumber   string `xorm:"serial_number"`                  //lpar.LogicalSerialNumber
	OSVersion      string `xorm:"os_version"`                     //lpar.OperatingSystemVersion
	Type           string `xorm:"type"`                           //lpar.PartitionType (LPAR/VIOServer
	PartitionState string `xorm:"-"`                              //lpar.PartitionState

	Location string `xorm:"location"`

	EnableHMCStats  bool `xorm:"'enable_hmc_stats' default 1"`
	EnableNmonStats bool `xorm:"'enable_nmon_stats' default 1"`

	NmonFreq     int    `xorm:"'nmon_freq' default 60" binding:"Default(60);IntegerNotZero"`
	NmonOutDB    string `xorm:"nmon_outdb"`
	NmonIP       string `xorm:"nmon_ip"`
	NmonSSHUser  string `xorm:"nmon_ssh_user"`
	NmonSSHKey   string `xorm:"nmon_ssh_key"`
	NmonLogLevel string `xorm:"'nmon_loglevel' default 'info'" binding:"Default(info)"`
	NmonFilePath string `xorm:"'nmon_filepath' default '/var/log/nmon/%{hostname}_%Y%m%d_%H%M.nmon'" binding:"Default(/var/log/nmon/%{hostname}_%Y%m%d_%H%M.nmon)"`

	ExtraTags []string `xorm:"extra-tags"` //common tags for nmon and also for hmc stats

	Description string `xorm:"description"`
}

// DBConfig read from DB
type DBConfig struct {
	Influxdb map[string]*InfluxCfg
	HMC      map[string]*HMCCfg
	Devices  map[string]*DeviceCfg
}

// Init initialices the DB
func Init(cfg *DBConfig) error {

	log.Debug("--------------------Initializing Config-------------------")

	log.Debug("-----------------------END Config metrics----------------------")
	return nil
}
