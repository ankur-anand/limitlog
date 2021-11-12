package limitlog_test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/ankur-anand/limitlog"
)

var (
	update = flag.Bool("update", false, "update the golden files of this test")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestMemLogger_Add_Search(t *testing.T) {
	ml := limitlog.NewInMemLogDB(2)
	ml.Add(1, "We need to manage logs on a system with limited memory.")
	ml.Add(2, "We need to query which of the logs contain a given word.")
	keys := ml.Search("We", 2)

	if !reflect.DeepEqual(keys, []int{2, 1}) {
		t.Errorf("Expected Keys Didn't match")
	}
	ml.Add(2, "The first line of the input is the maximum size of logs you should keep S.")
	keys = ml.Search("We", 2)

	if !reflect.DeepEqual(keys, []int{1}) {
		t.Errorf("Expected Keys Didn't match")
	}
	keys = ml.Search("Logs", 2)

	if !reflect.DeepEqual(keys, []int{2, 1}) {
		t.Errorf("Expected Keys Didn't match")
	}
	keys = ml.Search("Logs", 1)

	if !reflect.DeepEqual(keys, []int{2}) {
		t.Errorf("Expected Keys Didn't match")
	}

	// should evict the 1
	ml.Add(3, "The last line contains the single word END denoting the end of the program.")

	keys = ml.Search("Logs", 2)
	if !reflect.DeepEqual(keys, []int{2}) {
		t.Errorf("Expected Keys Didn't match")
	}

	keys = ml.Search("We", 2)
	if !reflect.DeepEqual(keys, []int{}) {
		t.Errorf("Expected Keys Didn't match")
	}

	keys = ml.Search("the", 2)
	if !reflect.DeepEqual(keys, []int{3, 2}) {
		t.Errorf("Expected Keys Didn't match")
	}

	ml.Add(1, "We need to manage logs on a system with limited memory.")
	ml.Add(2, "We need to query which of the logs contain a given word.")
	keys = ml.Search("We", 2)

	if !reflect.DeepEqual(keys, []int{2, 1}) {
		t.Errorf("Expected Keys Didn't match")
	}
}

func TestReaderInWriterOutLog_ReadFrom(t *testing.T) {
	inout := limitlog.ReaderInWriterOutLog{}

	rdr := &bytes.Buffer{}
	rdr.WriteString("2")
	rdr.WriteString("\n")
	rdr.WriteString("ADD 56 the first \n")
	rdr.WriteString("SEARCH the 1 \n")
	rdr.WriteString("ADD 25 the second log \n")
	rdr.WriteString("SEARCH the 2 \n")
	rdr.WriteString("\n") // some new line added for test
	rdr.WriteString("\n")
	rdr.WriteString("ADD 67 the third log \n")
	rdr.WriteString("SEARCH the 3 \n")
	rdr.WriteString("\n")
	rdr.WriteString("\n")
	rdr.WriteString("SEARCH fourth 1 \n")
	rdr.WriteString("\n")
	rdr.WriteString("\n")
	rdr.WriteString("END \n")

	resultWriter := &bytes.Buffer{}
	_, err := inout.ReadFrom(rdr, resultWriter)
	if err != nil {
		t.Error(err)
	}

	expected := &bytes.Buffer{}

	expected.WriteString("56\n")
	expected.WriteString("25 56\n")
	expected.WriteString("67 25\n")
	expected.WriteString("NONE\n")
	expected.WriteString("END")

	expectedBytes := expected.Bytes()

	resultBytes := resultWriter.Bytes()

	if !bytes.Equal(expectedBytes, resultBytes) {
		t.Errorf("expected result didn't matched with written result")
	}
}

func TestReaderInWriterOutLog_ReadFromGoldenFile(t *testing.T) {
	inout := limitlog.ReaderInWriterOutLog{}
	reader, l, err := goldenGenerate(t, "key100000", *update, 100000)
	if err != nil {
		t.Error(err)
	}

	resultWriter := &bytes.Buffer{}
	_, err = inout.ReadFrom(reader, resultWriter)
	if err != nil {
		t.Error(err)
	}

	if l > 0 {
		expected := &bytes.Buffer{}

		expected.WriteString("56\n")
		expected.WriteString("25 56\n")
		expected.WriteString("67 25 56\n")
		expected.WriteString("NONE\n")
		expected.WriteString("END")

		expectedBytes := expected.Bytes()

		resultBytes := resultWriter.Bytes()

		if !bytes.Equal(expectedBytes, resultBytes) {
			t.Errorf("expected result didn't matched with written result")
		}
	}
}

func BenchmarkSearch_FilledCapacity_2Keys(b *testing.B) {
	s := 2
	inMemDB := limitlog.NewInMemLogDB(s)
	for i := 0; i < s; i++ {
		key, _ := strconv.Atoi(randKeyOfLength(15))
		line := randString(15)
		inMemDB.Add(key, line)
	}

	inMemDB.Add(56, "the first")
	inMemDB.Add(25, "the second log")
	inMemDB.Add(67, "the third log")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := inMemDB.Search("the", 3)
		if !reflect.DeepEqual(res, []int{67, 25}) {
			b.Errorf("Expected Keys Didn't match")
		}
	}
}

func BenchmarkSearch_FilledCapacity_100KKeys(b *testing.B) {
	s := 100000
	inMemDB := limitlog.NewInMemLogDB(s)
	for i := 0; i < s; i++ {
		key, _ := strconv.Atoi(randKeyOfLength(15))
		line := randString(15)
		inMemDB.Add(key, line)
	}

	inMemDB.Add(56, "the first")
	inMemDB.Add(25, "the second log")
	inMemDB.Add(67, "the third log")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := inMemDB.Search("the", 3)
		if !reflect.DeepEqual(res, []int{67, 25, 56}) {
			b.Errorf("Expected Keys Didn't match")
		}
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
var numbers = []rune("0123456789")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randKeyOfLength(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = numbers[rand.Intn(len(numbers))]
	}
	return string(b)
}

func randWord(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randString(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s = s + " " + randWord(15)
	}
	return s
}

func goldenGenerate(t *testing.T, goldenFile string, update bool, s int) (io.Reader, int, error) {
	t.Helper()
	goldenPath := "testdata/" + goldenFile + ".golden"

	if update {
		buff := &bytes.Buffer{}
		buff.WriteString(fmt.Sprintf("%d \n", s))
		for i := 0; i < s-3; i++ {
			key := randKeyOfLength(15)
			line := randString(15)
			toWrite := fmt.Sprintf("ADD %s %s\n", key, line)
			buff.WriteString(toWrite)
		}
		buff.WriteString("ADD 56 the first \n")
		buff.WriteString("SEARCH the 1 \n")
		buff.WriteString("ADD 25 the second log \n")
		buff.WriteString("SEARCH the 2 \n")
		buff.WriteString("ADD 67 the third log \n")
		buff.WriteString("SEARCH the 3 \n")
		buff.WriteString("SEARCH fourth 1 \n")
		buff.WriteString("END \n")
		f, err := os.OpenFile(goldenPath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			t.Error(err)
		}
		n, err := f.Write(buff.Bytes())
		if err != nil {
			t.Errorf("error writing to file %v", err)
		}
		t.Logf("golden file updated wrote %d bytes", n)
		f.Close()
	}
	f, err := os.OpenFile(goldenPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {

		t.Error(err)
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("error reading the golden file %v", err)
	}
	return bytes.NewReader(content), len(content), nil
}
