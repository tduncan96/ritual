package db

type Host struct {
	HostId  int64  `db:"HostId"`
	Name    string `db:"Name"`
	Address string `db:"Address"`
	Port    int64  `db:"Port"`
	User    string `db:"User"`
	KeyPath string `db:"KeyPath"`
}

func GetHost(hostName string) (host Host, err error) {
	err = DB.Get(&host, "SELECT * FROM Hosts WHERE = ?", hostName)
	return host, err
}