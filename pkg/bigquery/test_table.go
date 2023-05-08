package bigquery

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
)

const testTableName = "junit"

var suites = []string{
	"openshift-tests",
	"openshift-tests-upgrade",
	"BakckendDisruption",
	"Cluster upgrade",
	"hypershift-e2e",
	"cluster install",
	"Operator results",
}

type TestTableManager struct {
	ctx     context.Context
	client  *Client
	dataset string
}

func NewTestTableManager(ctx context.Context, client *Client) *TestTableManager {
	return &TestTableManager{
		ctx:    ctx,
		client: client,
	}
}

func (tm *TestTableManager) ListTests() ([]v1.TestInfo, error) {
	now := time.Now()
	log.Infof("fetching unique test/suite names from bigquery")
	table := tm.client.bigquery.Dataset(tm.dataset).Table(testTableName)

	sql := fmt.Sprintf(`
		SELECT DISTINCT
		    test_name as name,
		    testsuite as suite
		FROM
			%s.%s.%s
		WHERE
		    testsuite IN ('%s')
		AND
		    test_name NOT LIKE 'steph graph.%%'
		AND
		    test_name NOT LIKE 'Run multi-stage test%%'
		AND
		    test_name NOT LIKE '%% was not OOMKilled%%'
		ORDER BY name, testsuite DESC`,
		table.ProjectID, tm.client.datasetName, table.TableID, strings.Join(suites, "','"))
	log.Debugf("query is %q", sql)

	q := tm.client.bigquery.Query(sql)
	it, err := q.Read(tm.ctx)
	if err != nil {
		return nil, err
	}

	var results []v1.TestInfo
	for {
		var testInfo v1.TestInfo
		err := it.Next(&testInfo)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, testInfo)
	}
	log.WithFields(log.Fields{
		"count": len(results),
	}).Infof("fetched unique test/suite names from bigquery in %v", time.Since(now))

	return results, nil
}
