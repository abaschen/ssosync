package internal

import (
	"testing"

	"github.com/awslabs/ssosync/internal/config"
)

// Benchmark tests for performance-critical functions

func BenchmarkIgnoreUser(b *testing.B) {
	s := &syncGSuite{
		cfg: &config.Config{
			IgnoreUsers: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ignoreUser("user2@example.com")
	}
}

func BenchmarkIgnoreUserLargeList(b *testing.B) {
	// Create a large ignore list to test performance
	ignoreList := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		ignoreList[i] = "user" + string(rune(i)) + "@example.com"
	}

	s := &syncGSuite{
		cfg: &config.Config{
			IgnoreUsers: ignoreList,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ignoreUser("user500@example.com")
	}
}

func BenchmarkIgnoreGroup(b *testing.B) {
	s := &syncGSuite{
		cfg: &config.Config{
			IgnoreGroups: []string{"group1@example.com", "group2@example.com", "group3@example.com"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ignoreGroup("group2@example.com")
	}
}

func BenchmarkIncludeGroup(b *testing.B) {
	s := &syncGSuite{
		cfg: &config.Config{
			IncludeGroups: []string{"group1@example.com", "group2@example.com", "group3@example.com"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.includeGroup("group2@example.com")
	}
}

// Test the performance improvement of our hash set optimization
func BenchmarkIgnoreUserHashSet(b *testing.B) {
	// Create a large ignore list
	ignoreList := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		ignoreList[i] = "user" + string(rune(i)) + "@example.com"
	}

	s := &syncGSuite{
		cfg: &config.Config{
			IgnoreUsers: ignoreList,
		},
	}

	// Pre-populate the hash set by calling ignoreUser once
	s.ignoreUser("user0@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ignoreUser("user500@example.com")
	}
}

// Benchmark the old linear search approach for comparison
func benchmarkIgnoreUserLinear(b *testing.B, ignoreList []string, searchUser string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		found := false
		for _, u := range ignoreList {
			if u == searchUser {
				found = true
				break
			}
		}
		_ = found
	}
}

func BenchmarkIgnoreUserLinearSmall(b *testing.B) {
	ignoreList := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	benchmarkIgnoreUserLinear(b, ignoreList, "user2@example.com")
}

func BenchmarkIgnoreUserLinearLarge(b *testing.B) {
	ignoreList := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		ignoreList[i] = "user" + string(rune(i)) + "@example.com"
	}
	benchmarkIgnoreUserLinear(b, ignoreList, "user500@example.com")
}
