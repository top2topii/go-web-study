// Package seed는 개발·데모용 이커머스 데이터를 생성한다.
// 고정 시드 난수를 쓰므로 실행할 때마다 같은 데이터가 나온다(날짜는 실행 시점 기준).
package seed

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"time"

	"shopadmin/internal/store"
)

type Options struct {
	Customers int
	Products  int
	Days      int // 주문 생성 기간(오늘부터 과거로)
	Now       time.Time
}

func DefaultOptions() Options {
	return Options{Customers: 200, Products: 120, Days: 180, Now: time.Now().UTC()}
}

type Result struct {
	Categories, Products, Customers, Orders, Items int
}

// Run은 하나의 트랜잭션 안에서 모든 데이터를 넣는다.
// 중간에 실패하면 Rollback되어 DB는 원래 상태 그대로다.
func Run(ctx context.Context, db *sql.DB, opt Options) (res Result, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}
	// 이름 붙은 반환값(err)을 defer에서 확인하는 관용구.
	// 성공(err == nil)했으면 Commit, 아니면 Rollback.
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	q := store.New(tx) // sqlc Queries는 *sql.DB와 *sql.Tx 모두 받는다(DBTX 인터페이스)
	r := rand.New(rand.NewPCG(42, 2026))

	catIDs, err := seedCategories(ctx, q)
	if err != nil {
		return res, fmt.Errorf("categories: %w", err)
	}
	res.Categories = len(catIDs)

	// products에는 주문 가능한 active 상품만 담긴다. 생성 개수는 opt.Products.
	products, err := seedProducts(ctx, q, r, catIDs, opt.Products)
	if err != nil {
		return res, fmt.Errorf("products: %w", err)
	}
	res.Products = opt.Products

	customers, err := seedCustomers(ctx, q, r, opt)
	if err != nil {
		return res, fmt.Errorf("customers: %w", err)
	}
	res.Customers = len(customers)

	res.Orders, res.Items, err = seedOrders(ctx, q, r, opt, customers, products)
	if err != nil {
		return res, fmt.Errorf("orders: %w", err)
	}
	return res, nil
}

// ---- categories ----

type catDef struct {
	name, slug string
	children   [][2]string // {name, slug}
}

var categoryTree = []catDef{
	{"전자기기", "electronics", [][2]string{{"스마트폰", "smartphones"}, {"노트북", "laptops"}, {"오디오", "audio"}, {"액세서리", "accessories"}}},
	{"패션", "fashion", [][2]string{{"상의", "tops"}, {"하의", "bottoms"}, {"신발", "shoes"}}},
	{"홈·리빙", "home", [][2]string{{"가구", "furniture"}, {"주방", "kitchen"}, {"조명", "lighting"}}},
	{"뷰티", "beauty", [][2]string{{"스킨케어", "skincare"}, {"메이크업", "makeup"}}},
	{"스포츠", "sports", [][2]string{{"피트니스", "fitness"}, {"아웃도어", "outdoor"}}},
}

// seedCategories는 하위 카테고리 ID만 반환한다(상품은 하위 카테고리에 속한다).
func seedCategories(ctx context.Context, q *store.Queries) ([]int64, error) {
	var leaves []int64
	for _, top := range categoryTree {
		parent, err := q.CreateCategory(ctx, store.CreateCategoryParams{Name: top.name, Slug: top.slug})
		if err != nil {
			return nil, err
		}
		for _, ch := range top.children {
			c, err := q.CreateCategory(ctx, store.CreateCategoryParams{
				ParentID: &parent.ID, // 포인터 필드: 값이 있으면 주소, NULL이면 nil
				Name:     ch[0],
				Slug:     ch[1],
			})
			if err != nil {
				return nil, err
			}
			leaves = append(leaves, c.ID)
		}
	}
	return leaves, nil
}

// ---- products ----

var (
	adjectives = []string{"프리미엄", "베이직", "슬림", "클래식", "울트라", "미니", "프로", "에센셜", "라이트", "빈티지"}
	nouns      = []string{"에디션", "컬렉션", "시리즈", "모델", "세트", "플러스", "맥스", "에어", "원", "X"}
)

type product struct {
	id    int64
	price int64
}

func seedProducts(ctx context.Context, q *store.Queries, r *rand.Rand, catIDs []int64, n int) ([]product, error) {
	out := make([]product, 0, n) // 길이 0, 용량 n: append 시 재할당 없음
	for i := 1; i <= n; i++ {
		cat := catIDs[r.IntN(len(catIDs))]
		// 가격: 5,000 ~ 약 500,000원, 로그 분포(저가 상품이 더 많게), 100원 단위
		price := int64(5000*math.Pow(100, r.Float64())) / 100 * 100

		status := "active"
		switch x := r.IntN(100); {
		case x < 10:
			status = "draft"
		case x < 18:
			status = "archived"
		}
		stock := int64(r.IntN(300))
		if r.IntN(10) == 0 {
			stock = 0 // 품절 상품 약 10%
		}

		p, err := q.CreateProduct(ctx, store.CreateProductParams{
			CategoryID:  &cat,
			Sku:         fmt.Sprintf("SKU-%05d", i),
			Name:        fmt.Sprintf("%s %s %d", adjectives[r.IntN(len(adjectives))], nouns[r.IntN(len(nouns))], i),
			Description: "데모용 상품 설명입니다.",
			Price:       price,
			Stock:       stock,
			Status:      status,
		})
		if err != nil {
			return nil, err
		}
		if status == "active" {
			out = append(out, product{id: p.ID, price: p.Price})
		}
	}
	return out, nil
}

// ---- customers ----

var (
	lastNames  = []string{"김", "이", "박", "최", "정", "강", "조", "윤", "장", "임", "한", "오", "서", "신", "권"}
	firstNames = []string{"민준", "서연", "도윤", "하은", "시우", "지우", "주원", "서윤", "예준", "지민", "하준", "수아", "지호", "채원", "현우"}
)

type customer struct {
	id      int64
	created time.Time
}

func seedCustomers(ctx context.Context, q *store.Queries, r *rand.Rand, opt Options) ([]customer, error) {
	out := make([]customer, 0, opt.Customers)
	for i := 1; i <= opt.Customers; i++ {
		// 가입일: 최근 1년 안에서 무작위
		created := opt.Now.Add(-time.Duration(r.Int64N(int64(365 * 24 * time.Hour))))
		status := "active"
		switch x := r.IntN(100); {
		case x < 8:
			status = "inactive"
		case x < 10:
			status = "banned"
		}
		var phone *string
		if r.IntN(4) != 0 {
			s := fmt.Sprintf("010-%04d-%04d", r.IntN(10000), r.IntN(10000))
			phone = &s
		}
		c, err := q.CreateCustomer(ctx, store.CreateCustomerParams{
			Email:     fmt.Sprintf("customer%03d@example.com", i),
			Name:      lastNames[r.IntN(len(lastNames))] + firstNames[r.IntN(len(firstNames))],
			Phone:     phone,
			Status:    status,
			CreatedAt: created,
		})
		if err != nil {
			return nil, err
		}
		out = append(out, customer{id: c.ID, created: c.CreatedAt})
	}
	// 가입일 순 정렬: 주문일 이전에 가입한 고객만 빠르게 고르기 위해
	sort.Slice(out, func(i, j int) bool { return out[i].created.Before(out[j].created) })
	return out, nil
}

// ---- orders ----

func seedOrders(ctx context.Context, q *store.Queries, r *rand.Rand, opt Options, customers []customer, products []product) (orders, items int, err error) {
	for day := opt.Days - 1; day >= 0; day-- {
		date := opt.Now.AddDate(0, 0, -day).Truncate(24 * time.Hour)

		// 하루 주문 수: 기본 4건 + 기간에 따라 증가하는 추세 + 주말 가중치 + 노이즈
		progress := 1 - float64(day)/float64(opt.Days)
		base := 4 + 8*progress
		if wd := date.Weekday(); wd == time.Saturday || wd == time.Sunday {
			base *= 1.4
		}
		n := int(base + r.NormFloat64()*2)
		if n < 0 {
			n = 0
		}

		for range n { // Go 1.22+: 정수 range
			at := date.Add(time.Duration(r.Int64N(int64(24 * time.Hour))))
			if at.After(opt.Now) {
				continue
			}
			// sort.Search: 가입일이 at 이후인 첫 고객 인덱스 → 그 앞까지가 후보
			k := sort.Search(len(customers), func(i int) bool { return customers[i].created.After(at) })
			if k == 0 {
				continue
			}
			cust := customers[r.IntN(k)]

			// 주문 항목 1~4개
			type line struct {
				p   product
				qty int64
			}
			lines := make([]line, 1+r.IntN(4))
			var total int64
			for i := range lines {
				lines[i] = line{p: products[r.IntN(len(products))], qty: int64(1 + r.IntN(3))}
				total += lines[i].p.price * lines[i].qty
			}

			o, err := q.CreateOrder(ctx, store.CreateOrderParams{
				CustomerID: cust.id,
				Status:     orderStatus(r, opt.Now.Sub(at)),
				Total:      total,
				CreatedAt:  at,
				UpdatedAt:  at,
			})
			if err != nil {
				return orders, items, err
			}
			for _, l := range lines {
				if err := q.CreateOrderItem(ctx, store.CreateOrderItemParams{
					OrderID: o.ID, ProductID: l.p.id, Quantity: l.qty, UnitPrice: l.p.price,
				}); err != nil {
					return orders, items, err
				}
				items++
			}
			orders++
		}
	}
	return orders, items, nil
}

// orderStatus는 주문 경과 시간에 따라 그럴듯한 상태를 고른다.
func orderStatus(r *rand.Rand, age time.Duration) string {
	x := r.IntN(100)
	switch {
	case x < 4:
		return "cancelled"
	case x < 7 && age > 72*time.Hour:
		return "refunded"
	}
	switch {
	case age < 6*time.Hour:
		return "pending"
	case age < 24*time.Hour:
		return "paid"
	case age < 72*time.Hour:
		return "shipped"
	default:
		return "delivered"
	}
}
