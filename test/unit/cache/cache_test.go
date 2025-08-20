package cache_test

import (
	"L0/internal/cache"
	"L0/test/testutils"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoCache_BasicOperations(t *testing.T) {
	// Arrange
	c := cache.New(5*time.Minute, 10*time.Minute)
	testOrder := testutils.MinimalOrderFixture("test_order")

	t.Run("set_and_get", func(t *testing.T) {
		// Act
		c.Set("test_key", testOrder, 0)
		value, found := c.Get("test_key")

		// Assert
		assert.True(t, found)
		require.NotNil(t, value)
		// Проверяем, что значение правильно сохранилось
		assert.Equal(t, testOrder, value)
	})

	t.Run("get_nonexistent", func(t *testing.T) {
		// Act
		value, found := c.Get("nonexistent_key")

		// Assert
		assert.False(t, found)
		assert.Nil(t, value)
	})

	t.Run("delete", func(t *testing.T) {
		// Arrange
		c.Set("delete_test", testOrder, 0)

		// Act
		c.Delete("delete_test")
		value, found := c.Get("delete_test")

		// Assert
		assert.False(t, found)
		assert.Nil(t, value)
	})

	t.Run("overwrite", func(t *testing.T) {
		// Arrange
		firstOrder := testutils.MinimalOrderFixture("first")
		secondOrder := testutils.MinimalOrderFixture("second")

		// Act
		c.Set("overwrite_key", firstOrder, 0)
		c.Set("overwrite_key", secondOrder, 0)
		value, found := c.Get("overwrite_key")

		// Assert
		assert.True(t, found)
		require.NotNil(t, value)
		assert.Equal(t, secondOrder, value)
	})
}

func TestGoCache_TTL(t *testing.T) {
	c := cache.New(100*time.Millisecond, 50*time.Millisecond) // Короткие TTL для тестов
	testOrder := testutils.MinimalOrderFixture("ttl_test")

	t.Run("item_expires", func(t *testing.T) {
		// Arrange
		c.Set("ttl_key", testOrder, 50*time.Millisecond)

		// Act - проверяем сразу
		value, found := c.Get("ttl_key")
		assert.True(t, found)
		assert.Equal(t, testOrder, value)

		// Wait for expiration
		time.Sleep(100 * time.Millisecond)

		// Assert - должно исчезнуть
		value, found = c.Get("ttl_key")
		assert.False(t, found)
		assert.Nil(t, value)
	})

	t.Run("default_ttl", func(t *testing.T) {
		// Arrange
		c.Set("default_ttl_key", testOrder, 0) // 0 означает использование default TTL

		// Act - проверяем сразу
		value, found := c.Get("default_ttl_key")
		assert.True(t, found)
		assert.Equal(t, testOrder, value)

		// Wait for default expiration (100ms)
		time.Sleep(150 * time.Millisecond)

		// Assert - должно исчезнуть
		value, found = c.Get("default_ttl_key")
		assert.False(t, found)
		assert.Nil(t, value)
	})
}

func TestGoCache_ConcurrentAccess(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)

	t.Run("concurrent_writes", func(t *testing.T) {
		const numGoroutines = 100
		var wg sync.WaitGroup

		// Act - несколько горутин записывают разные значения
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				order := testutils.MinimalOrderFixture("order_" + string(rune(id)))
				c.Set("concurrent_key", order, 0)
			}(i)
		}

		wg.Wait()

		// Assert - кеш не должен паниковать, значение должно быть одним из записанных
		value, found := c.Get("concurrent_key")
		assert.True(t, found)
		assert.NotNil(t, value)
	})

	t.Run("concurrent_read_write", func(t *testing.T) {
		const numReaders = 50
		const numWriters = 50
		var wg sync.WaitGroup

		// Подготавливаем данные
		initialOrder := testutils.MinimalOrderFixture("initial")
		c.Set("rw_key", initialOrder, 0)

		results := make(chan bool, numReaders)

		// Readers
		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, found := c.Get("rw_key")
				results <- found
			}()
		}

		// Writers
		for i := 0; i < numWriters; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				order := testutils.MinimalOrderFixture("writer_" + string(rune(id)))
				c.Set("rw_key", order, 0)
			}(i)
		}

		wg.Wait()
		close(results)

		// Assert - все читатели должны получить какое-то значение (true или false)
		readCount := 0
		for found := range results {
			if found {
				readCount++
			}
		}

		// Должны быть успешные чтения
		assert.Greater(t, readCount, 0)
	})

	t.Run("concurrent_delete", func(t *testing.T) {
		const numGoroutines = 50
		var wg sync.WaitGroup

		// Подготавливаем данные
		for i := 0; i < numGoroutines; i++ {
			order := testutils.MinimalOrderFixture("delete_" + string(rune(i)))
			c.Set("delete_key_"+string(rune(i)), order, 0)
		}

		// Act - удаляем параллельно
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				c.Delete("delete_key_" + string(rune(id)))
			}(i)
		}

		wg.Wait()

		// Assert - все ключи должны быть удалены
		for i := 0; i < numGoroutines; i++ {
			_, found := c.Get("delete_key_" + string(rune(i)))
			assert.False(t, found)
		}
	})
}

func TestGoCache_DataTypes(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)

	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"string", "string_key", "test_string"},
		{"int", "int_key", 42},
		{"struct", "struct_key", testutils.MinimalOrderFixture("struct_test")},
		{"slice", "slice_key", []string{"a", "b", "c"}},
		{"map", "map_key", map[string]int{"one": 1, "two": 2}},
		{"nil", "nil_key", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			c.Set(tt.key, tt.value, 0)
			value, found := c.Get(tt.key)

			// Assert
			assert.True(t, found)
			assert.Equal(t, tt.value, value)
		})
	}
}

func TestGoCache_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	c := cache.New(5*time.Minute, 10*time.Minute)
	const numOperations = 10000

	t.Run("write_performance", func(t *testing.T) {
		order := testutils.MinimalOrderFixture("perf_test")

		start := time.Now()
		for i := 0; i < numOperations; i++ {
			c.Set("perf_key_"+string(rune(i)), order, 0)
		}
		duration := time.Since(start)

		t.Logf("Write performance: %d operations in %v (%.2f ops/ms)",
			numOperations, duration, float64(numOperations)/float64(duration.Milliseconds()))

		assert.Less(t, duration, 1*time.Second)
	})

	t.Run("read_performance", func(t *testing.T) {
		// Подготавливаем данные
		order := testutils.MinimalOrderFixture("read_perf")
		for i := 0; i < 1000; i++ {
			c.Set("read_key_"+string(rune(i)), order, 0)
		}

		start := time.Now()
		for i := 0; i < numOperations; i++ {
			c.Get("read_key_" + string(rune(i%1000)))
		}
		duration := time.Since(start)

		t.Logf("Read performance: %d operations in %v (%.2f ops/ms)",
			numOperations, duration, float64(numOperations)/float64(duration.Milliseconds()))

		assert.Less(t, duration, 500*time.Millisecond)
	})
}
