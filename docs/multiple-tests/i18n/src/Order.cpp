#include <iostream>
#include <string>
#include <vector>
#include <ctime>
#include <iomanip>   // for number formatting

struct Order {
    int id;
    std::string customer;
    int quantity;
    std::string status;
    double price;
};

class OrderManager {
    std::vector<Order> orders;
    int nextId = 1;

public:
    void createOrder(const std::string& customer, int qty, double price) {
        Order o{nextId++, customer, qty, "NEW", price};
        orders.push_back(o);

        // ❌ Hardcoded success message
        std::cout << "Order created successfully for customer: " 
                  << customer << " with quantity " << qty 
                  << " and price " << price << std::endl;
    }

    void listOrders() {
        std::cout << "------ Order List ------" << std::endl; // ❌ Hardcoded label

        for (auto& o : orders) {
            std::cout << "Order ID: " << o.id << ", "
                      << "Customer: " << o.customer << ", "
                      << "Qty: " << o.quantity << ", "
                      // ❌ Hardcoded status mapping
                      << "Status: " << (o.status == "NEW" ? "New Order" : o.status) << ", "
                      // ❌ Locale-unaware currency formatting
                      << "Price: $" << std::fixed << std::setprecision(2) << o.price
                      << std::endl;
        }

        std::cout << "------ End of Orders ------" << std::endl; // ❌ Hardcoded footer
    }

    void deleteOrder(int id) {
        for (auto it = orders.begin(); it != orders.end(); ++it) {
            if (it->id == id) {
                orders.erase(it);
                // ❌ Hardcoded delete confirmation
                std::cout << "Order deleted successfully!" << std::endl;
                return;
            }
        }
        // ❌ Hardcoded error message
        std::cout << "Error: Order not found." << std::endl;
    }

    void printReport() {
        // ❌ Locale-unaware date formatting (fixed US-style format)
        std::time_t now = std::time(nullptr);
        char buffer[80];
        std::strftime(buffer, sizeof(buffer), "%m/%d/%Y %H:%M:%S", std::localtime(&now));
        std::cout << "Report generated at: " << buffer << std::endl;

        // ❌ Hardcoded label + locale-unaware number formatting
        double revenue = 0;
        for (auto& o : orders) {
            revenue += o.price * o.quantity;
        }

        std::cout << "Total Orders: " << orders.size() << std::endl;
        std::cout << "Total Revenue: " << revenue << std::endl; // ❌ Missing locale formatting
    }
};

int main() {
    OrderManager manager;

    manager.createOrder("Alice", 3, 1234.56);
    manager.createOrder("Bob", 5, 98765.43);

    manager.listOrders();

    manager.deleteOrder(2);
    manager.deleteOrder(10); // should print error

    manager.printReport();

    return 0;
}
