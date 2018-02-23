package hmcpcm

//-----------------------------------------------------------
// XML parsing structures
//-----------------------------------------------------------

/*--------------------------------------
* XML MANAGE
*--------------------------------------*/

// Feed base struct of Atom feed
type Feed struct {
	//XMLName xml.Name `xml:"feed"`
	Entries []Entry `xml:"entry"`
}

// Entry is the atom feed section containing the links to PCM data and the Category
type Entry struct {
	//XMLName xml.Name `xml:"entry"`
	ID   string `xml:"id"`
	Link struct {
		Href string `xml:"href,attr"`
		Type string `xml:"type,attr"`
	} `xml:"link,omitempty"`
	Contents []Content `xml:"content"`
	Category struct {
		Term string `xml:"term,attr"`
	} `xml:"category,omitempty"`
}

// Content feed struct containing all managed systems
type Content struct {
	//XMLName xml.Name        `xml:"content"`
	System []ManagedSystem `xml:"http://www.ibm.com/xmlns/systems/power/firmware/uom/mc/2012_10/ ManagedSystem"`
}

// ManagedSystem struct contains a managed system and his associated partitions
type ManagedSystem struct {
	//XMLName                     xml.Name `xml:"http://www.ibm.com/xmlns/systems/power/firmware/uom/mc/2012_10/ ManagedSystem"`
	SystemName                  string
	State                       string
	AssociatedLogicalPartitions Partitions `xml:"http://www.ibm.com/xmlns/systems/power/firmware/uom/mc/2012_10/ AssociatedLogicalPartitions" json:"-"`
	AssociatedVirtualIOServers  Partitions `xml:"http://www.ibm.com/xmlns/systems/power/firmware/uom/mc/2012_10/ AssociatedVirtualIOServers" json:"-"`
	//Only
	Lpars map[string]*LogicalPartition `xml:"-"`
	Vios  map[string]*VirtualIOServer  `xml:"-"`
	UUID  string                       `xml:"-"`
}

// Partitions contains links to the partition informations
type Partitions struct {
	Links []Link `xml:"link,omitempty"`
}

// Link the link itself is stored in the attribute href
type Link struct {
	Href string `xml:"href,attr"`
}

/*--------------------------------------
* XML LPAR
*--------------------------------------*/

// LparEntry is the atom feed section containing the LPAR info
type LparEntry struct {
	//XMLName  xml.Name      `xml:"entry"`
	ID       string        `xml:"id"`
	Contents []LparContent `xml:"content"`
}

// LparContent feed struct containing all managed systems
type LparContent struct {
	//XMLName xml.Name           `xml:"content"`
	Lpar []LogicalPartition `xml:"http://www.ibm.com/xmlns/systems/power/firmware/uom/mc/2012_10/ LogicalPartition"`
}

//LogicalPartition Contains
type LogicalPartition struct {
	//XMLName                xml.Name `xml:"LogicalPartition"`
	LogicalSerialNumber    string `xml:"LogicalSerialNumber"`
	OperatingSystemVersion string `xml:"OperatingSystemVersion"`
	PartitionName          string `xml:"PartitionName"`
	PartitionState         string `xml:"PartitionState"`
	PartitionType          string `xml:"PartitionType"`
	PartitionUUID          string `xml:"PartitionUUID"`
}

/*--------------------------------------
* XML LPAR
*--------------------------------------*/

// ViosEntry is the atom feed section containing the LPAR info
type ViosEntry struct {
	//XMLName  xml.Name      `xml:"entry"`
	ID       string        `xml:"id"`
	Contents []ViosContent `xml:"content"`
}

// VirtualIOServer info for virtual servers
type VirtualIOServer LogicalPartition

// ViosContent feed struct containing all managed systems
type ViosContent struct {
	//XMLName xml.Name           `xml:"content"`
	Vios []VirtualIOServer `xml:"http://www.ibm.com/xmlns/systems/power/firmware/uom/mc/2012_10/ VirtualIOServer"`
}
