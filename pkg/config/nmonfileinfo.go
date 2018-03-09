package config

import "fmt"

/***************************
	Influx DB backends
	-GetNmonFileInfoCfgByID(struct)
	-GetNmonFileInfoMap (map - for interna config use
	-GetNmonFileInfoArray(Array - for web ui use )
	-AddNmonFileInfo
	-DelNmonFileInfo
	-UpdateNmonFileInfo
  -GetNmonFileInfoAffectOnDel
***********************************/

/*GetNmonFileInfoByIDFile get device data by id*/
func (dbc *DatabaseCfg) GetNmonFileInfoByIDFile(id string, filename string) (NmonFileInfo, error) {
	cfgarray, err := dbc.GetNmonFileInfoArray("id='" + id + "' and file_name='" + filename + "'")
	if err != nil {
		return NmonFileInfo{}, err
	}
	if len(cfgarray) > 1 {
		return NmonFileInfo{}, fmt.Errorf("Error %d results on get NmonFileInfo by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return NmonFileInfo{}, fmt.Errorf("Error no values have been returned with this id %s in the influx config table", id)
	}
	return *cfgarray[0], nil
}

/*GetNmonFileInfoByID get device data by id*/
func (dbc *DatabaseCfg) GetNmonFileInfoByID(id string) (NmonFileInfo, error) {
	cfgarray, err := dbc.GetNmonFileInfoArray("id='" + id + "'")
	if err != nil {
		return NmonFileInfo{}, err
	}
	if len(cfgarray) > 1 {
		return NmonFileInfo{}, fmt.Errorf("Error %d results on get NmonFileInfo by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return NmonFileInfo{}, fmt.Errorf("Error no values have been returned with this id %s in the influx config table", id)
	}
	return *cfgarray[0], nil
}

/*GetNmonFileInfoMap  return data in map format*/
func (dbc *DatabaseCfg) GetNmonFileInfoMap(filter string) (map[string]*NmonFileInfo, error) {
	cfgarray, err := dbc.GetNmonFileInfoArray(filter)
	cfgmap := make(map[string]*NmonFileInfo)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetNmonFileInfoArray generate an array of devices with all its information */
func (dbc *DatabaseCfg) GetNmonFileInfoArray(filter string) ([]*NmonFileInfo, error) {
	var err error
	var devices []*NmonFileInfo
	//Get Only data for selected devices
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get NmonFileInfo  data filteter with %s : %v\n", filter, err)
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

/*AddNmonFileInfo for adding new devices*/
func (dbc *DatabaseCfg) AddNmonFileInfo(dev NmonFileInfo) (int64, error) {
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
	log.Infof("Added new influx backend Successfully with id %s ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelNmonFileInfo for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelNmonFileInfo(id string) (int64, error) {
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

	affected, err = session.Where("id='" + id + "'").Delete(&NmonFileInfo{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully influx db with ID %s [ %d Devices Affected  ]", id, affecteddev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

// AddOrUpdateNmonFileInfo this method insert data if not previouosly exist the tuple ifxServer.Name or update it if already exist
func (dbc *DatabaseCfg) AddOrUpdateNmonFileInfo(dev *NmonFileInfo) (int64, error) {
	log.Debugf("ADD OR UPDATE %+v", dev)
	//check if exist
	m, err := dbc.GetNmonFileInfoArray("id == '" + dev.ID + "'")
	if err != nil {
		return 0, err
	}
	switch len(m) {
	case 1:
		log.Debugf("Updating Device %+v", m)
		return dbc.UpdateNmonFileInfo(m[0].ID, *dev)
	case 0:
		log.Debugf("Adding new Device %+v", dev)
		return dbc.AddNmonFileInfo(*dev)
	default:
		log.Errorf("There is some error when searching for db %+v , found %d", dev, len(m))
		return 0, fmt.Errorf("There is some error when searching for db %+v , found %d", dev, len(m))
	}

}

/*UpdateNmonFileInfo for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateNmonFileInfo(id string, dev NmonFileInfo) (int64, error) {
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
		log.Infof("Updated Influx Config to %s devices ", affecteddev)
	}

	affected, err = session.Where("id='" + id + "'").UseBool().AllCols().Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Infof("Updated Influx Config Successfully with id %s and data:%+v, affected", id, dev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*GetNmonFileInfoAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetNmonFileInfoAffectOnDel(id string) ([]*DbObjAction, error) {
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
