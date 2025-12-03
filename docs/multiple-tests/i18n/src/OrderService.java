import java.text.MessageFormat;
import java.util.ResourceBundle;

public class OrderService {
    private final ResourceBundle messages;

    public OrderService(ResourceBundle messages) {
        this.messages = messages;
    }

    public void processOrder(String customer, int count) {
        // BAD: Concatenated string
        System.out.println("Processing order for " + customer + " with " + count + " items."); // ❌

        // GOOD: Localized
        String msg = messages.getString("order.processing");
        System.out.println(MessageFormat.format(msg, customer, count));

        if (count > 2) {
            String success = messages.getString("order.success");
            System.out.println(MessageFormat.format(success, customer));
        } else {
            // BAD: Hardcoded success
            System.out.println("Order placed successfully!"); // ❌
        }
    }
}
