import i18n from "i18next";
import { initReactI18next } from "react-i18next";

const resources = {
  en: {
    translation: {
      "order.title": "Order Management",
      "order.create": "Create Order",
      "order.customer": "Customer",
      "order.date": "Order Date",
      "order.amount": "Amount",
      "order.success": "Order created successfully",
      "order.error": "Failed to create order",
      "order.total": "Total Orders: {{count}}",
      "order.revenue": "Total Revenue: {{revenue, number}}",
    },
  },
  fr: {
    translation: {
      "order.title": "Gestion des commandes",
      // missing "order.success" -> should fallback
      "order.create": "Créer une commande",
    },
  },
  pseudo: {
    translation: new Proxy({}, {
      get: (_, key) => `[[${key}]]`, // Pseudo-localization
    }),
  },
};

i18n.use(initReactI18next).init({
  resources,
  lng: "en", // default
  fallbackLng: "en", // fallback
  debug: true,
  interpolation: {
    escapeValue: false,
    format: function (value, format) {
      if (format === "number") {
        return new Intl.NumberFormat(i18n.language).format(value);
      }
      if (format === "date") {
        return new Intl.DateTimeFormat(i18n.language).format(value);
      }
      return value;
    },
  },
});

export default i18n;
