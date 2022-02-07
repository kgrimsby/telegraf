package snmp

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestTrapLookup(t *testing.T) {
	tests := []struct {
		name     string
		oid      string
		expected MibEntry
	}{
		{
			name: "Known trap OID",
			oid:  ".1.3.6.1.6.3.1.1.5.1",
			expected: MibEntry{
				MibName: "TGTEST-MIB",
				OidText: "coldStart",
			},
		},
		{
			name: "Known trap value OID",
			oid:  ".1.3.6.1.2.1.1.3.0",
			expected: MibEntry{
				MibName: "TGTEST-MIB",
				OidText: "sysUpTimeInstance",
			},
		},
		{
			name: "Unknown enterprise sub-OID",
			oid:  ".1.3.6.1.4.1.0.1.2.3",
			expected: MibEntry{
				MibName: "TGTEST-MIB",
				OidText: "enterprises.0.1.2.3",
			},
		},
		{
			name:     "Unknown MIB",
			oid:      ".1.2.3",
			expected: MibEntry{OidText: "iso.2.3"},
		},
	}

	// Load the MIBs
	require.NoError(t, LoadMibsFromPath([]string{"testdata/mibs"}, testutil.Logger{}, &GosmiMibLoader{}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the actual test
			actual, err := TrapLookup(tt.oid)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestTrapLookupFail(t *testing.T) {
	tests := []struct {
		name     string
		oid      string
		expected string
	}{
		{
			name:     "New top level OID",
			oid:      ".3.6.1.3.0",
			expected: "Could not find node for OID 3.6.1.3.0",
		},
		{
			name:     "Malformed OID",
			oid:      ".1.3.dod.1.3.0",
			expected: "could not convert OID .1.3.dod.1.3.0: strconv.ParseUint: parsing \"dod\": invalid syntax",
		},
	}

	// Load the MIBs
	require.NoError(t, LoadMibsFromPath([]string{"testdata/mibs"}, testutil.Logger{}, &GosmiMibLoader{}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the actual test
			_, err := TrapLookup(tt.oid)
			require.EqualError(t, err, tt.expected)
		})
	}
}

type TestingMibLoader struct {
	folders []string
	files   []string
}

func (t *TestingMibLoader) appendPath(path string) {
	t.folders = append(t.folders, path)
}

func (t *TestingMibLoader) loadModule(path string) error {
	t.files = append(t.files, path)
	return nil
}
func TestFolderLookup(t *testing.T) {
	paths := [][]string{
		{"testdata"},
		{"testdata", "loadMibsFromPath"},
		{"testdata", "loadMibsFromPath", "linkTarget"},
		{"testdata", "loadMibsFromPath", "root"},
		{"testdata", "mibs"},
	}
	tests := []struct {
		name    string
		folders []string
		files   []string
	}{
		{
			name:  "loading folders",
			files: []string{"emptyFile", "testmib"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := TestingMibLoader{}

			err := LoadMibsFromPath([]string{"testdata"}, testutil.Logger{}, &loader)
			require.NoError(t, err)
			for _, pathSlice := range paths {
				path := filepath.Join(pathSlice...)
				tt.folders = append(tt.folders, path)
			}
			require.Equal(t, tt.folders, loader.folders)
			if runtime.GOOS == "windows" {
				tt.files = []string{"emptyFile", "symlink", "testmib"}
			}
			require.Equal(t, tt.files, loader.files)
		})
	}
}
