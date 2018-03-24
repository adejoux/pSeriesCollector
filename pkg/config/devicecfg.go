package config

import "fmt"

/***************************
	Device DB backends
	-GetDeviceCfgCfgByID(struct)
	-GetDeviceCfgMap (map - for interna config use
	-GetDeviceCfgArray(Array - for web ui use )
	-AddDeviceCfg
	-DelDeviceCfg
	-UpdateDeviceCfg
  -GetDeviceCfgAffectOnDel
***********************************/

/*GetDeviceCfgByID get device data by id*/
func (dbc *DatabaseCfg) GetDeviceCfgByID(id string) (DeviceCfg, error) {
	cfgarray, err := dbc.GetDeviceCfgArray("id='" + id + "'")
	if err != nil {
		return DeviceCfg{}, err
	}
	if len(cfgarray) > 1 {
		return DeviceCfg{}, fmt.Errorf("Error %d results on get DeviceCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return DeviceCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the Devices config table", id)
	}
	return *cfgarray[0], nil
}

/*GetDeviceCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetDeviceCfgMap(filter string) (map[string]*DeviceCfg, error) {
	cfgarray, err := dbc.GetDeviceCfgArray(filter)
	cfgmap := make(map[string]*DeviceCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetDeviceCfgArray generate an array of devices with all its information */
func (dbc *DatabaseCfg) GetDeviceCfgArray(filter string) ([]*DeviceCfg, error) {
	var err error
	var devices []*DeviceCfg
	//Get Only data for selected devices
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get DeviceCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get influxcfg   data: %v\n", err)
			return nil, err
		}
	}
	return devices, nil
}

/*AddDeviceCfg for adding new devices*/
func (dbc *DatabaseCfg) AddDeviceCfg(dev DeviceCfg) (int64, error) {
	var err error
	var affected int64
	session := dbc.x.NewSession()
	defer session.Close()

	affected, err = session.Insert(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	//no other relation
	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Added new Devices backend Successfully with id %s ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelDeviceCfg for deleting Devices databases from ID*/
func (dbc *DatabaseCfg) DelDeviceCfg(id string) (int64, error) {
	var affecteddev, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in HMCCfg

	affecteddev, err = session.Where("outdb='" + id + "'").Cols("outdb").Update(&HMCCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Device with id on delete HMCCfg with id: %s, error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&DeviceCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Devices db with ID %s [ %d Devices Affected  ]", id, affecteddev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

// AddOrUpdateDeviceCfg this method insert data if not previouosly exist the tuple ifxServer.Name or update it if already exist
func (dbc *DatabaseCfg) AddOrUpdateDeviceCfg(dev *DeviceCfg) (int64, error) {
	log.Debugf("ADD OR UPDATE %+v", dev)
	//check if exist
	m, err := dbc.GetDeviceCfgArray("id == '" + dev.ID + "'")
	if err != nil {
		return 0, err
	}
	switch len(m) {
	case 1:
		log.Debugf("Updating Device ID[%s] %#+v", m[0].ID, dev)
		return dbc.UpdateDeviceCfgBase(m[0].ID, *dev)
	case 0:
		dev.EnableHMCStats = true
		dev.EnableNmonStats = true
		log.Debugf("Adding new Device ID[%s] %#+v", dev.ID, dev)
		return dbc.AddDeviceCfg(*dev)
	default:
		log.Errorf("There is some error when searching for db %+v , found %d", dev, len(m))
		return 0, fmt.Errorf("There is some error when searching for db %+v , found %d", dev, len(m))
	}

}

/*UpdateDeviceCfgBase for adding new Devices from Scanned products , don't change current configurations*/
func (dbc *DatabaseCfg) UpdateDeviceCfgBase(id string, dev DeviceCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()

	affected, err = session.Where("id='" + id + "'").Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Infof("Updated Device  Base Config Successfully with id %s and data:%+v, affected", id, dev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*UpdateDeviceCfg for adding new Devices*/
func (dbc *DatabaseCfg) UpdateDeviceCfg(id string, dev DeviceCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()
	if id != dev.ID { //ID has been changed
		affecteddev, err = session.Where("outdb='" + id + "'").Cols("outdb").Update(&HMCCfg{OutDB: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error on Update InfluxConfig on update id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Infof("Updated Device Config to %s devices ", affecteddev)
	}
	log.Debugf("Updating Device %#+v", dev)
	affected, err = session.Where("id='" + id + "'").UseBool().AllCols().Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Infof("Updated Device Config Successfully with id %s and data:%+v, affected", id, dev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*GetDeviceCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetDeviceCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var devices []*HMCCfg
	var obj []*DbObjAction
	if err := dbc.x.Where("outdb='" + id + "'").Find(&devices); err != nil {
		log.Warnf("Error on Get Outout db id %d for devices , error: %s", id, err)
		return nil, err
	}

	for _, val := range devices {
		obj = append(obj, &DbObjAction{
			Type:     "HMCCfg",
			TypeDesc: "HMC Devices",
			ObID:     val.ID,
			Action:   "Reset InfluxDB Server from HMC  Device to 'default' InfluxDB Server",
		})

	}
	return obj, nil
}
