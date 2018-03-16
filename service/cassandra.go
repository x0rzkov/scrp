package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gocql/gocql"
)

// Cassandra struct is the primary data structure for the plugin
type Cassandra struct {
	// URL is only for backwards compatibility
	URL              string
	URLs             []string `toml:"urls"`
	Username         string
	Password         string
	Keyspace         string `toml:"keyspace"`
	UserAgent        string
	RetentionPolicy  string
	WriteConsistency string
	UDPPayload       int               `toml:"udp_payload"`
	HTTPProxy        string            `toml:"http_proxy"`
	HTTPHeaders      map[string]string `toml:"http_headers"`
	ContentEncoding  string            `toml:"content_encoding"`

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	VerifyHost bool `toml:"verify_host"`

	Retry bool `toml:"retry"`

	// Precision is only here for legacy support. It will be ignored.
	Precision string

	session *gocql.Session
}

// Connect initiates the primary connection to the range of provided URLs
func (i *Cassandra) Connect() error {
	var urls []string
	urls = append(urls, i.URLs...)
	cluster := gocql.NewCluster(i.URLs...)
	cluster.Keyspace = i.Keyspace
	cluster.Consistency = gocql.Quorum
	if i.SSLCA != "" {
		sslOpts := &gocql.SslOptions{
			CaPath:                 i.SSLCA,
			EnableHostVerification: i.VerifyHost,
		}
		if i.SSLCert != "" && i.SSLKey != "" {
			sslOpts.CertPath = i.SSLCert
			sslOpts.KeyPath = i.SSLKey
		}
		cluster.SslOpts = sslOpts
	}

	i.session, _ = cluster.CreateSession()

	rand.Seed(time.Now().UnixNano())
	return nil
}

// Close will terminate the session to the backend, returning error if an issue arises
func (i *Cassandra) Close() error {
	if !i.session.Closed() {
		i.session.Close()
	}
	return nil
}

func (i *Cassandra) Check() error {
	// UPDATE users
	// SET email = ‘janedoe@abc.com’
	// WHERE login = 'jdoe'
	// IF email = ‘jdoe@abc.com’;

	// 	BEGIN BATCH
	//   INSERT INTO purchases (user, balance) VALUES ('user1', -8) IF NOT EXISTS;
	//   INSERT INTO purchases (user, expense_id, amount, description, paid)
	//     VALUES ('user1', 1, 8, 'burrito', false);
	// APPLY BATCH;
	return nil
}

// Write will choose a random server in the cluster to write to until a successful write
// occurs, logging each unsuccessful. If all servers fail, return error.
func (i *Cassandra) Write(records map[string]string) error {
	//TODO: performance test against batching
	//fmt.Fprintf(os.Stderr, "Input packet", metrics)
	// This will get set to nil if a successful write occurs
	err := fmt.Errorf("Could not write to any cassandra server in cluster")
	counters := make(map[string]int)
	// regexCount, _ := regexp.Compile(`\.count\.(.*)`)
	// regexUpdate, _ := regexp.Compile(`\.update\.(.*)`)
	//insertBatch := i.session.NewBatch(gocql.UnloggedBatch)
	// for k, v := range records {
	// 	//fmt.Println("%s", tags) //Debugging only
	// 	if regexCount.MatchString(records["name"]) {
	// 		counter := regexCount.FindStringSubmatch(records["name"])[1]
	// 		counters[counter] = counters[counter] + 1
	// 	} else if regexUpdate.MatchString(tags["name"]) && tags["msg"] != "" {
	// 		timestamp := time.Now().UTC()
	// 		if tags["updated"] != "" {
	// 			millis, err := strconv.ParseInt(tags["updated"], 10, 64)
	// 			if err == nil {
	// 				timestamp = time.Unix(0, millis*int64(time.Millisecond))
	// 			}
	// 		}
	// 		if rowError := i.session.Query(`INSERT INTO updates (id, updated, msg) values (?,?,?)`,
	// 			regexUpdate.FindStringSubmatch(tags["name"])[1],
	// 			timestamp,
	// 			tags["msg"]).Exec(); rowError != nil {
	// 			err = rowError //And let it continue
	// 		} else {
	// 			err = nil
	// 		}
	// 	} else {
	// 		if tags["id"] == "" {
	// 			tags["id"] = gocql.TimeUUID().String()
	// 		}
	// 		serialized, _ := json.Marshal(tags)
	// 		//insertBatch.Query(`INSERT INTO logs JSON ?`, string(serialized))
	// 		if rowError := i.session.Query(`INSERT INTO logs JSON ?`, string(serialized)).Exec(); rowError != nil {
	// 			err = rowError //And let it continue
	// 		} else {
	// 			err = nil
	// 		}
	// 	}
	// }

	for key, value := range counters {
		if rowError := i.session.Query(`UPDATE counters set total=total+? where id=?;`, value, key).Exec(); rowError != nil {
			err = rowError //And let it continue
		} else {
			err = nil
		}
	}

	//err = i.session.ExecuteBatch(insertBatch)
	if !i.Retry && err != nil {
		fmt.Fprintf(os.Stderr, "!E CASSANDRA OUTPUT PLUGIN - NOT RETRYING %s", err.Error())
		err = nil //Do not retry
	}
	return err
}
