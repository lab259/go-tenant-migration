package mtnt_test

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/lab259/go-migration"
	"github.com/lab259/rlog"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	migrationRLog "github.com/lab259/go-migration/rlog"

	mtnt "github.com/lab259/go-migration-tenant"
	"github.com/lab259/go-migration-tenant/mongo"
	mtntRLog "github.com/lab259/go-migration-tenant/rlog"
)

type FakeAccount struct {
	db *mgo.Database
}

func (account *FakeAccount) Identification() string {
	return account.db.Name
}

func (account *FakeAccount) ProvideMigrationContext(fnc func(executionContext interface{}) error) error {
	return fnc(account.db)
}

type FakeAccountProducer struct {
	session *mgo.Session
	hasNext bool
}

func (producer *FakeAccountProducer) Get() ([]mtnt.Account, error) {
	return []mtnt.Account{
		&FakeAccount{
			db: producer.session.DB("database1"),
		},
		&FakeAccount{
			db: producer.session.DB("database2"),
		},
		&FakeAccount{
			db: producer.session.DB("database3"),
		},
		&FakeAccount{
			db: producer.session.DB("database4"),
		},
	}, nil
}

func (producer *FakeAccountProducer) HasNext() bool {
	hn := producer.hasNext
	producer.hasNext = false
	return hn
}

func (producer *FakeAccountProducer) Total() int {
	return 4
}

var _ = Describe("MigrationExecutor", func() {
	var session *mgo.Session

	BeforeEach(func() {
		sess, err := mgo.DialWithTimeout("mongodb://localhost:27017", time.Second*3)
		Expect(err).ToNot(HaveOccurred())
		session = sess
		dbs, err := session.DatabaseNames()
		Expect(err).ToNot(HaveOccurred())
		for _, db := range dbs {
			switch db {
			case "admin", "config", "local", "test":
				continue
			}
			Expect(session.DB(db).DropDatabase()).To(Succeed())
		}
	})

	It("should migrate with 1 worker", func() {
		source := migration.NewCodeSource()

		execution := make([]string, 0)
		source.Register(migration.NewMigration(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC), "Migration 1", func(executionContext interface{}) error {
			var ec *mgo.Database
			Expect(executionContext).To(BeAssignableToTypeOf(ec))
			execution = append(execution, "1")
			return nil
		}))
		source.Register(migration.NewMigration(time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC), "Migration 2", func(executionContext interface{}) error {
			var ec *mgo.Database
			Expect(executionContext).To(BeAssignableToTypeOf(ec))
			execution = append(execution, "2")
			return nil
		}))
		source.Register(migration.NewMigration(time.Date(2019, 1, 3, 0, 0, 0, 0, time.UTC), "Migration 3", func(executionContext interface{}) error {
			var ec *mgo.Database
			Expect(executionContext).To(BeAssignableToTypeOf(ec))
			execution = append(execution, "3")
			return nil
		}))
		logger := rlog.WithFields(nil)
		executor := mtnt.NewMigrationExecutor(migrationRLog.NewRLogReporter(logger, func(i int) {
			Fail("exit should not be called")
		}), mtntRLog.NewRLogReporter(logger), mongo.Connector, &FakeAccountProducer{
			session: session,
			hasNext: true,
		}, source)
		executor.Run(1, "migrate")

		dbs, err := session.DatabaseNames()
		Expect(err).ToNot(HaveOccurred())
		Expect(dbs).To(ContainElement("database1"))
		Expect(dbs).To(ContainElement("database2"))
		Expect(dbs).To(ContainElement("database3"))
		Expect(dbs).To(ContainElement("database4"))
		Expect(execution).To(HaveLen(12))
		Expect(execution).To(ConsistOf("1", "2", "3", "1", "2", "3", "1", "2", "3", "1", "2", "3"))
	})

	It("should migrate with 3 worker", func() {
		source := migration.NewCodeSource()

		execution := make([]string, 0)
		source.Register(migration.NewMigration(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC), "Migration 1", func(executionContext interface{}) error {
			var ec *mgo.Database
			Expect(executionContext).To(BeAssignableToTypeOf(ec))
			execution = append(execution, "1")
			return nil
		}))
		source.Register(migration.NewMigration(time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC), "Migration 2", func(executionContext interface{}) error {
			var ec *mgo.Database
			Expect(executionContext).To(BeAssignableToTypeOf(ec))
			execution = append(execution, "2")
			return nil
		}))
		source.Register(migration.NewMigration(time.Date(2019, 1, 3, 0, 0, 0, 0, time.UTC), "Migration 3", func(executionContext interface{}) error {
			var ec *mgo.Database
			Expect(executionContext).To(BeAssignableToTypeOf(ec))
			execution = append(execution, "3")
			return nil
		}))
		logger := rlog.WithFields(nil)
		executor := mtnt.NewMigrationExecutor(migrationRLog.NewRLogReporter(logger, func(i int) {
			Fail("exit should not be called")
		}), mtntRLog.NewRLogReporter(logger), mongo.Connector, &FakeAccountProducer{
			session: session,
			hasNext: true,
		}, source)
		executor.Run(3, "migrate")

		dbs, err := session.DatabaseNames()
		Expect(err).ToNot(HaveOccurred())
		Expect(dbs).To(ContainElement("database1"))
		Expect(dbs).To(ContainElement("database2"))
		Expect(dbs).To(ContainElement("database3"))
		Expect(dbs).To(ContainElement("database4"))
		Expect(execution).To(HaveLen(12))
		Expect(execution).To(ConsistOf("1", "2", "3", "1", "2", "3", "1", "2", "3", "1", "2", "3"))
	})
})