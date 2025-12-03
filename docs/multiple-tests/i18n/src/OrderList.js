// OrderList.js (excerpt with intentional i18n gaps)

import React, { useEffect, useState } from "react";

export default function OrderList() {
  const [orders, setOrders] = useState([]);

  useEffect(() => {
    fetch("http://localhost:8080/api/orders")
      .then((res) => res.json())
      .then(setOrders);
  }, []);

  return (
    <div>
      {/* ❌ Hardcoded title */}
      <h1>Order Management</h1>  

      <button onClick={() => alert("Create Order Clicked!")}>
        {/* ❌ Hardcoded label */}
        Create Order  
      </button>

      <table border="1">
        <thead>
          <tr>
            {/* ❌ Hardcoded column headers */}
            <th>Customer Name</th>
            <th>Quantity</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          {orders.map((o) => (
            <tr key={o.id}>
              <td>{o.customerName}</td>
              <td>{o.quantity}</td>
              {/* ❌ Hardcoded status mapping */}
              <td>{o.status === "NEW" ? "New Order" : o.status}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
