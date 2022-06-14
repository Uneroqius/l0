package store

import (
	"fmt"

	_ "github.com/lib/pq"

	"database/sql"

	"l0/internal/order"
)

type Cache map[string]order.Order

func (c *Cache) addOrder(order *order.Order) error {
	(*c)[order.OrderUID] = *order

	return nil
}

func (c *Cache) getOrderByID(order_uid string) (order *order.Order, isExist bool) {
	order_, isExist := (*c)[order_uid]
	if isExist {
		order = &order_
	}

	return
}

type Store struct {
	Cache  *Cache
	Config *Config
	db     *sql.DB
}

func (s *Store) loadDBToCache() (err error) {
	rows, err := s.db.Query(`
		SELECT 
			Orders.order_uid,
			Orders.track_number,
			Orders.entry,
			Delivery."name",
			Delivery.phone,
			Delivery.zip,
			Delivery.city,
			Delivery.address,
			Delivery.region,
			Delivery.email,
			Payments.transaction_,
			Payments.request_id ,
			Payments.currency,
			Payments.provider ,
			Payments.amount ,
			Payments.payment_dt ,
			Payments.bank ,
			Payments.delivery_cost ,
			Payments.goods_total ,
			Payments.custom_fee ,
			Orders.locale,
			Orders.internal_signature,
			Orders.customer_id,
			Orders.delivery_service,
			Orders.shardkey,
			Orders.sm_id,
			Orders.date_created,
			Orders.oof_shard
		FROM Orders
		INNER JOIN Delivery ON Orders.delivery_id  = Delivery.delivery_id
		INNER JOIN Payments ON Orders.payment_id = Payments.payment_id;
	`)
	if err != nil {
		return
	}

	for rows.Next() {
		ord := order.Order{
			Items: []order.Item{},
		}

		rows.Scan(
			&ord.OrderUID,
			&ord.TrackNumber,
			&ord.Entry,
			// Delivery
			&ord.Delivery.Name,
			&ord.Delivery.Phone,
			&ord.Delivery.Zip,
			&ord.Delivery.City,
			&ord.Delivery.Address,
			&ord.Delivery.Region,
			&ord.Delivery.Email,
			// Payment
			&ord.Payment.Transaction,
			&ord.Payment.RequestID,
			&ord.Payment.Currency,
			&ord.Payment.Provider,
			&ord.Payment.Amount,
			&ord.Payment.PaymentDt,
			&ord.Payment.Bank,
			&ord.Payment.DeliveryCost,
			&ord.Payment.GoodsTotal,
			&ord.Payment.CustomFee,
			//
			&ord.Locale,
			&ord.InternalSignature,
			&ord.CustomerID,
			&ord.DeliveryService,
			&ord.ShardKey,
			&ord.SmID,
			&ord.DateCreated,
			&ord.OofShard,
		)

		itemsRows, err := s.db.Query(`
			SELECT
				chrt_id,
				track_number,
				price,
				rid,
				name,
				sale,
				size,
				total_price,
				nm_id,
				brand,
				status
			FROM Items WHERE order_uid = $1;
		`, ord.OrderUID)

		if err != nil {
			return err
		}

		for itemsRows.Next() {
			item := order.Item{}

			itemsRows.Scan(
				&item.ChrtID,
				&item.TrackNumber,
				&item.Price,
				&item.RID,
				&item.Name,
				&item.Sale,
				&item.Size,
				&item.TotalPrice,
				&item.NmID,
				&item.Brand,
				&item.Status,
			)

			ord.Items = append(ord.Items, item)
		}

		s.Cache.addOrder(&ord)
	}

	return nil
}
func (s *Store) Open() error {
	db, err := sql.Open("postgres", s.Config.DatabaseURL)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	s.db = db

	err = s.loadDBToCache()
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) addDeliveryToDB(delivery *order.Delivery) (inserted_id int64, err error) {
	err = s.db.QueryRow(
		`INSERT INTO Delivery(name, phone, zip, city, address, region, email)
		 VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING delivery_id;`,
		delivery.Name,
		delivery.Phone,
		delivery.Zip,
		delivery.City,
		delivery.Address,
		delivery.Region,
		delivery.Email,
	).Scan(&inserted_id)

	return
}

func (s *Store) addPaymentToDB(p *order.Payment) (inserted_id int64, err error) {
	err = s.db.QueryRow(
		`INSERT INTO Payments(transaction_, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		 VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING payment_id;`,
		p.Transaction,
		p.RequestID,
		p.Currency,
		p.Provider,
		p.Amount,
		p.PaymentDt,
		p.Bank,
		p.DeliveryCost,
		p.GoodsTotal,
		p.CustomFee,
	).Scan(&inserted_id)

	return
}

func (s *Store) addOrderToDB(o *order.Order, DeliveryID, PaymentID int64) (err error) {
	_, err = s.db.Exec(
		`INSERT INTO Orders(order_uid, track_number, entry, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		 VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`,
		o.OrderUID,
		o.TrackNumber,
		o.Entry,
		DeliveryID,
		PaymentID,
		o.Locale,
		o.InternalSignature,
		o.CustomerID,
		o.DeliveryService,
		o.ShardKey,
		o.SmID,
		o.DateCreated,
		o.OofShard,
	)

	return
}

func (s *Store) addItemsToDB(i *[]order.Item, order_uid string) (err error) {
	items := *i

	for i := range items {
		if _, err = s.db.Exec(
			`INSERT INTO Items(order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			 VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`,
			order_uid,
			items[i].ChrtID,
			items[i].TrackNumber,
			items[i].Price,
			items[i].RID,
			items[i].Name,
			items[i].Sale,
			items[i].Size,
			items[i].TotalPrice,
			items[i].NmID,
			items[i].Brand,
			items[i].Status,
		); err != nil {
			return
		}
	}

	return
}

func (s *Store) isAlreadyExsistsInBD(orderUID string) (isExists bool) {
	_, isExists = s.Cache.getOrderByID(orderUID)
	return
}

func (s *Store) AddOrderToBD(order *order.Order) error {
	if s.isAlreadyExsistsInBD(order.OrderUID) {
		return fmt.Errorf("order with uid %s is already exists", order.OrderUID)
	}

	if err := s.addOrderToCache(order); err != nil {
		return err
	}

	if deliveryID, err := s.addDeliveryToDB(&order.Delivery); err != nil {
		return err
	} else if paymentID, err := s.addPaymentToDB(&order.Payment); err != nil {
		return err
	} else if err := s.addOrderToDB(order, deliveryID, paymentID); err != nil {
		return err
	} else if err := s.addItemsToDB(&order.Items, order.OrderUID); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetOrderByID(order_uid string) (order *order.Order, err error) {
	order, isExist := s.Cache.getOrderByID(order_uid)
	if !isExist {
		err = fmt.Errorf("order with uid %s is not exist in database", order_uid)
	}

	return
}

func (s *Store) addOrderToCache(order *order.Order) error {
	s.Cache.addOrder(order)

	return nil
}

func New(config *Config) *Store {
	return &Store{
		Config: config,
		Cache:  &Cache{},
	}
}
