package main

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
)

var incr int64 = 0
var incrPtr = &incr

type order struct {
	user_id, order_id, amount, price int64
	side                             bool // true for buy, false for sell
}

type orderBook struct {
	BuyOrders, SellOrders []order
}

type balance struct {
	user_id         int64
	assets          map[string]int64
	activeOrders    []order
	completedOrders []order
}

type users []balance

func (users *users) ModifyBalance(user_id int64, asset string, volume int64) {
	for _, u := range *users {
		if u.user_id == user_id {
			u.assets[asset] += volume
			break
		}
	}
	new_user := balance{user_id: user_id, assets: map[string]int64{asset: volume}}
	*users = append(*users, new_user)
}

func (ob *orderBook) DelOrder(order order) {
	if order.side { //del from buy orders
		if len(ob.BuyOrders) == 1 {
			ob.BuyOrders = ob.BuyOrders[0:0]
		} else {
			for i, v := range ob.BuyOrders {
				if v.order_id == order.order_id {
					ob.BuyOrders = append(ob.BuyOrders[:i], ob.BuyOrders[i+1:]...)
				}
			}
		}
	} else { //del from sell orders
		if len(ob.SellOrders) == 1 {
			ob.SellOrders = ob.SellOrders[0:0]
		} else {
			for i, v := range ob.SellOrders {
				if v.order_id == order.order_id {
					ob.SellOrders = append(ob.SellOrders[:i], ob.SellOrders[i+1:]...)
				}
			}
		}
	}
}

func (order *order) MatchOrders(ob *orderBook, users *users) {
	completed_orders := 0
	if order.side { //matching buy order
		for i, sellOrd := range ob.SellOrders {
			if sellOrd.price <= order.price {
				if order.amount < sellOrd.amount {
					remaining_amount := sellOrd.amount - order.amount
					total_price := sellOrd.price * order.amount
					for _, u := range *users { // changing assets
						if u.user_id == order.user_id {
							u.assets["UAH"] += order.amount
							u.assets["USD"] -= total_price
							u.completedOrders = append(u.completedOrders, *order)
						}
						if u.user_id == sellOrd.user_id {
							u.assets["USD"] += total_price
							u.assets["UAH"] -= order.amount
						}
					}
					sellOrd.amount = remaining_amount
					fmt.Printf("Buy order complete: bought %v UAH for %v USD each.\n", order.amount, sellOrd.price)
					ob.SellOrders[i-completed_orders] = sellOrd
					ob.DelOrder(*order)
					ob.print()
					break
				} else if order.amount == sellOrd.amount {
					total_price := sellOrd.price * order.amount
					for _, u := range *users { // changing assets
						if u.user_id == order.user_id {
							u.assets["UAH"] += order.amount
							u.assets["USD"] -= total_price
							u.completedOrders = append(u.completedOrders, *order)
						}
						if u.user_id == sellOrd.user_id {
							u.assets["USD"] += total_price
							u.assets["UAH"] -= order.amount
							u.completedOrders = append(u.completedOrders, *order)
						}
					}
					fmt.Printf("Buy order complete: bought %v UAH for %v USD each.\n", order.amount, sellOrd.price)
					fmt.Printf("Sell order complete: sold %v UAH for %v USD each.\n", order.amount, sellOrd.price)
					ob.DelOrder(*order)
					ob.DelOrder(sellOrd)
					ob.print()
					break
				} else {
					total_price := sellOrd.price * sellOrd.amount
					remaining_amount := order.amount - sellOrd.amount
					for _, u := range *users { // changing assets
						if u.user_id == order.user_id {
							u.assets["UAH"] += sellOrd.amount
							u.assets["USD"] -= total_price
						}
						if u.user_id == sellOrd.user_id {
							u.assets["USD"] += total_price
							u.assets["UAH"] -= sellOrd.amount
							u.completedOrders = append(u.completedOrders, sellOrd)
						}
					}
					order.amount = remaining_amount
					for i, ord := range ob.BuyOrders {
						if ord.order_id == ord.order_id {
							ob.BuyOrders[i] = *order
						}
					}
					fmt.Printf("Sell order complete: sold %v UAH for %v USD each.\n", sellOrd.amount, sellOrd.price)
					ob.DelOrder(sellOrd)
					completed_orders++
					ob.print()
				}
			}
		}
	} else { //matching sell order
		for i, buyOrd := range ob.BuyOrders {
			if buyOrd.price >= order.price {
				if order.amount < buyOrd.amount {
					remaining_amount := buyOrd.amount - order.amount
					total_price := buyOrd.price * order.amount
					for _, u := range *users { // changing assets
						if u.user_id == order.user_id {
							u.assets["UAH"] -= order.amount
							u.assets["USD"] += total_price
							u.completedOrders = append(u.completedOrders, *order)
						}
						if u.user_id == buyOrd.user_id {
							u.assets["USD"] -= total_price
							u.assets["UAH"] += order.amount
						}
					}
					buyOrd.amount = remaining_amount
					ob.BuyOrders[i-completed_orders] = buyOrd
					fmt.Printf("Sell order complete: sold %v UAH for %v USD each.\n", order.amount, buyOrd.price)
					ob.DelOrder(*order)
					completed_orders++
					ob.print()
					break
				} else if order.amount == buyOrd.amount {
					total_price := buyOrd.price * order.amount
					for _, u := range *users { // changing assets
						if u.user_id == order.user_id {
							u.assets["UAH"] -= order.amount
							u.assets["USD"] += total_price
							u.completedOrders = append(u.completedOrders, *order)
						}
						if u.user_id == buyOrd.user_id {
							u.assets["USD"] -= total_price
							u.assets["UAH"] += order.amount
							u.completedOrders = append(u.completedOrders, *order)
						}
					}
					fmt.Printf("Sell order complete: sold %v UAH for %v USD each.\n", order.amount, buyOrd.price)
					fmt.Printf("Buy order complete: bought %v UAH for %v USD each.\n", order.amount, buyOrd.price)
					ob.DelOrder(*order)
					ob.DelOrder(buyOrd)
					ob.print()
					break
				} else {
					remaining_amount := order.amount - buyOrd.amount
					total_price := buyOrd.price * buyOrd.amount
					for _, u := range *users { // changing assets
						if u.user_id == order.user_id {
							u.assets["UAH"] -= buyOrd.amount
							u.assets["USD"] += total_price
						}
						if u.user_id == buyOrd.user_id {
							u.assets["USD"] -= total_price
							u.assets["UAH"] += buyOrd.amount
							u.completedOrders = append(u.completedOrders, *order)
						}
					}
					order.amount = remaining_amount
					for i, ord := range ob.SellOrders {
						if ord.order_id == ord.order_id {
							ob.SellOrders[i] = *order
						}
					}
					fmt.Printf("Buy order complete: bought %v UAH for %v USD each.\n", buyOrd.amount, buyOrd.price)
					ob.DelOrder(buyOrd)
					ob.print()
				}
			}
		}
	}
}

func (ob *orderBook) AddOrder(order order, users *users) {
	order.order_id = *incrPtr
	enough_assets := false
	for _, v := range *users {
		if v.user_id == order.user_id {
			if order.side == true {
				if order.amount*order.price <= v.assets["USD"] {
					enough_assets = true
					break
				} else {
					fmt.Printf("Error: user (user_id: %v) doesn't have enough assets.\n", v.user_id)
					break
				}
			} else {
				if order.amount <= v.assets["UAH"] {
					enough_assets = true
					break
				} else {
					fmt.Printf("Error: user (user_id: %v) doesn't have enough assets.\n", v.user_id)
					break
				}
			}
		}
	}
	if enough_assets {
		if order.side {
			flag := false
			for i, v := range *users {
				if v.user_id == order.user_id {
					ob.BuyOrders = append(ob.BuyOrders, order)
					sort.SliceStable(ob.BuyOrders, func(i, j int) bool {
						return ob.BuyOrders[i].price > ob.BuyOrders[j].price
					})
					(*users)[i].activeOrders = append((*users)[i].activeOrders, order)
					*incrPtr++
					fmt.Printf("Added buy order: order_id - %v, user_id - %v, amount - %v, price - %v.\n", order.order_id, order.user_id, order.amount, order.price)
					flag = true
					break
				}
			}
			if !flag {
				fmt.Println("Error, there is no user with such user_id. Can't add order to order book.")
			}
		} else {
			flag := false
			for i, v := range *users {
				if v.user_id == order.user_id {
					ob.SellOrders = append(ob.SellOrders, order)
					sort.SliceStable(ob.SellOrders, func(i, j int) bool {
						return ob.SellOrders[i].price < ob.SellOrders[j].price
					})
					(*users)[i].activeOrders = append((*users)[i].activeOrders, order)
					*incrPtr++
					fmt.Printf("Added sell order: order_id - %v, user_id - %v, amount - %v, price - %v.\n", order.order_id, order.user_id, order.amount, order.price)
					flag = true
					break
				}
			}
			if !flag {
				fmt.Println("Error, there is no user with such user_id. Can't add order to order book.")
			}
		}
		ob.print()
		order.MatchOrders(ob, users)
	}
}

func (ob *orderBook) print() {
	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	fmt.Fprintln(writer, "Size\t\tBid\t\tAsk\t\tSize")
	if diff := len(ob.BuyOrders) - len(ob.SellOrders); diff <= 0 {
		ind := 0
		if len(ob.BuyOrders) != 0 {
			for i, order := range ob.BuyOrders {
				fmt.Fprintf(writer, "%v\t\t%v\t\t%v\t\t%v\n", order.amount, order.price, ob.SellOrders[i].price, ob.SellOrders[i].amount)
				ind = i
			}
			for j := ind + 1; j < len(ob.SellOrders); j++ {
				fmt.Fprintf(writer, "\t\t\t\t%v\t\t%v\n", ob.SellOrders[j].price, ob.SellOrders[j].amount)
			}
		} else {
			for i := 0; i < len(ob.SellOrders); i++ {
				fmt.Fprintf(writer, "\t\t\t\t%v\t\t%v\n", ob.SellOrders[i].price, ob.SellOrders[i].amount)
			}
		}
	} else {
		ind := 0
		if len(ob.SellOrders) != 0 {
			for i, order := range ob.SellOrders {
				fmt.Fprintf(writer, "%v\t\t%v\t\t%v\t\t%v\n", ob.BuyOrders[i].amount, ob.BuyOrders[i].price, order.price, order.amount)
				ind = i
			}
			for j := ind + 1; j < len(ob.BuyOrders); j++ {
				fmt.Fprintf(writer, "%v\t\t%v\n", ob.BuyOrders[j].amount, ob.BuyOrders[j].price)
			}
		} else {
			for i := 0; i < len(ob.BuyOrders); i++ {
				fmt.Fprintf(writer, "%v\t\t%v\n", ob.BuyOrders[i].amount, ob.BuyOrders[i].price)
			}
		}
	}
	writer.Flush()
}

func (b *balance) print() {
	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	fmt.Fprintf(writer, "User_id (%v) assets:\t|USD - %v|\t|UAH - %v|\n", b.user_id, b.assets["USD"], b.assets["UAH"])
	writer.Flush()
}

func main() {
	orderB := orderBook{}
	users := users{}
	users.ModifyBalance(1, "USD", 100)
	users.ModifyBalance(2, "USD", 100)
	users.ModifyBalance(3, "UAH", 100)
	users[1].print()
	users[2].print()
	orderB.AddOrder(order{user_id: 1, amount: 9, price: 18, side: true}, &users)
	orderB.AddOrder(order{user_id: 2, amount: 10, price: 6, side: true}, &users)
	orderB.AddOrder(order{user_id: 3, amount: 15, price: 5, side: false}, &users)
	orderB.AddOrder(order{user_id: 3, amount: 7, price: 6, side: true}, &users)
	users[1].print()
	users[2].print()
	//orderB.DelOrder(orderB.BuyOrders[0])
	//orderB.AddOrder(order{user_id: 4, amount: 500, price: 25, side: false}, &users)
}
