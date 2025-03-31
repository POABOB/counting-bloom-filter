# Counting Bloom Filter

`Counting Bloom Filter` 是一個 `使用計數的 Bloom Filter`，它可以 `儲存元素` 並支援 `元素的過期處理`。
這個過濾器適用於需要 `高效存儲大量資料`、`查詢是否存在的場景`，並且 `允許對已過期的元素進行清理`。

## 1. 使用方法

### 安裝

```bash
go get github.com/POABOB/counting-bloom-filter
```

### 創建 Counting Bloom Filter

你可以使用 `NewCountingBloomFilter` 或 `NewDefaultCountingBloomFilter` 來創建一個新的 Counting Bloom Filter。

#### 使用自定義大小的 Counting Bloom Filter 與過期策略

```go
package main

import (
	"fmt"
	"time"
	bloom "github.com/POABOB/counting-bloom-filter"
)

func main() {
	// 創建一個大小為 1MB 的 Counting Bloom Filter，並且每 30 秒定期刪除 1/10 的元素
	cbf := bloom.NewCountingBloomFilter(1 * 1024 * 1024, bloom.WithExpiryDuration(bloom.LAZY_EXPIRATION, 30 * time.Second))

	// 添加元素
	cbf.Add("item1")
	cbf.Add("item2")

	// 檢查元素是否存在
	fmt.Println(cbf.Check("item1")) // true
	fmt.Println(cbf.Check("item2")) // true
	fmt.Println(cbf.Check("item3")) // false
}
```

#### 使用預設的 Counting Bloom Filter

```go
package main

import (
	"fmt"
	bloom "github.com/POABOB/counting-bloom-filter"
)

func main() {
	// 使用預設的大小和設置創建 Counting Bloom Filter
	cbf := bloom.NewDefaultCountingBloomFilter()

	// 添加元素
	cbf.Add("item1")
	cbf.Add("item2")

	// 檢查元素是否存在
	fmt.Println(cbf.Check("item1")) // true
	fmt.Println(cbf.Check("item2")) // true
	fmt.Println(cbf.Check("item3")) // false
}
```

### 方法說明

- `Add(item string)`：將元素 `item` 添加到 Counting Bloom Filter 中。
- `Check(item string) bool`：檢查元素 `item` 是否在 Bloom Filter 中。若元素已過期或不存在，將返回 `false`。
- `Remove(item string)`：從 Bloom Filter 中移除元素 `item`。
- `RemoveAll()`：從 Bloom Filter 中移除所有元素。

## 2. 評估假陽性命中率

Bloom Filter 的 `假陽性命中率 (False Positive Rate, FPR)` 是指過濾器 `錯誤地判斷元素存在的機率`。根據設計，FPR 會受到以下因素的影響：
- `Bloom Filter 的 bit 大小 m`。
- `哈希函數的次數 K`。
- `插入的元素數量 n`。

根據這些因素，我們可以計算 FPR。假設：

- 使用 `12` 次哈希函數（`K = 12`）。
- 假設平均元素數量是 `33,333`。
- bit 大小為 `1MB`（即 `1024 * 1024` 位元）。

在這種配置下，布隆過濾器的 FPR 可以通過以下公式計算：

```text
FPR ≈ (1 - e^(-K * n / m))^K
```

在這種配置下，經過計算，FPR 約為 `0.00000165`（即 `1.65e-06`），這代表在平均插入 33,333 個元素的情況下，FPR 非常低。

所以，可以透過 `n (平均插入元素數量)` 來調整 `m (bit 大小)`，以達到最佳的性能和精確度。

如果希望達到 `1.65e-06` 的錯誤率且 `平均元素有 100 萬個`，建議 `m = 1_000_000 * 30`，`m` 設定為 `30MiB` 會比較好。

## 3. 預設值
- `m`：預設大小為 `1MB`（`1024 * 1024` 位元），這在存儲大約 33,333 個元素的情況下，FPR 為 `0.00000165`。
- `K`：預設使用 `12` 次哈希函數，這可以保證即使在高並發的情況下，FPR 依然保持在合理範圍。

## 4. 配置選項

你可以在創建 `CountingBloomFilter` 時傳入自定義選項來控制過濾器的行為，例如過期策略、清理間隔時間等。

### 支援的過期策略：
- `NO_EXPIRATION`：不過期。
- `LAZY_EXPIRATION`：模擬延遲過期（通過減少計數器來實現過期），每次會 `減少 1/10 的元素`。
- `RESET_EVERY_PERIOD`：定期 `重設所有元素` 的計數器。
- `EXPIRY_DURATION`：基於時間的過期策略，`當元素過期時將被移除`，記憶體使用量較大，`GB` 級別資料量不建議使用。

例如，若你想設置自定義的過期策略和清理時間間隔，可以這樣配置：
```go
package main

import (
	bloom "github.com/POABOB/counting-bloom-filter"
)

opts := bloom.WithExpiryDuration(bloom.LAZY_EXPIRATION, 60 * time.Second))
cbf := bloom_filter.NewCountingBloomFilter(1 * 1024 * 1024, opts)
```

