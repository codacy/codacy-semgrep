import React, { useState } from "react";
import { useTranslation } from "react-i18next";

export default function OrderList() {
  const { t, i18n } = useTranslation();
  const [orders, setOrders] = useState([
    { id: 1, customer: "Alice", amount: 1234.56, date: new Date() },
    { id: 2, customer: "Bob", amount: 98765.43, date: new Date() },
  ]);

  const switchLang = (lng) => i18n.changeLanguage(lng);

  const addOrder = () => {
    // ❌ BAD: Hardcoded + concatenation (gap)
    alert("Order for " + "Charlie" + " created on " + new Date());

    // ✅ GOOD: Use i18n key + interpolation
    // alert(t("order.success"));
  };

  return (
    <div>
      <h1>{t("order.title")}</h1>
      <div>
        <button onClick={() => switchLang("en")}>EN</button>
        <button onClick={() => switchLang("fr")}>FR</button>
        <button onClick={() => switchLang("pseudo")}>Pseudo</button>
      </div>

      <button onClick={addOrder}>{t("order.create")}</button>

      <table border="1" style={{ marginTop: 20 }}>
        <thead>
          <tr>
            <th>{t("order.customer")}</th>
            <th>{t("order.date")}</th>
            <th>{t("order.amount")}</th>
          </tr>
        </thead>
        <tbody>
          {orders.map((o) => (
            <tr key={o.id}>
              <td>{o.customer}</td>

              {/* ❌ BAD: Hardcoded date formatting */}
              <td>{o.date.toLocaleDateString("en-US")}</td>

              {/* ✅ GOOD: i18n aware date */}
              {/* <td>{t("order.date", { date: o.date, format: "date" })}</td> */}

              {/* ❌ BAD: Hardcoded number */}
              <td>${o.amount.toFixed(2)}</td>

              {/* ✅ GOOD: Locale-sensitive */}
              {/* <td>{t("order.amount", { amount: o.amount, format: "number" })}</td> */}
            </tr>
          ))}
        </tbody>
      </table>

      <p>{t("order.total", { count: orders.length })}</p>

      {/* ❌ BAD: Revenue displayed without locale formatting */}
      <p>Total Revenue: {orders.reduce((sum, o) => sum + o.amount, 0)}</p>

      {/* ✅ GOOD: Revenue with i18n interpolation */}
      {/* <p>{t("order.revenue", { revenue: orders.reduce((s, o) => s + o.amount, 0), format: "number" })}</p> */}
    </div>
  );
}
