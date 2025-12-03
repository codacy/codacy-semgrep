import java.text.MessageFormat;
import java.util.ResourceBundle;

public class PaymentService {
    private final ResourceBundle messages;

    public PaymentService(ResourceBundle messages) {
        this.messages = messages;
    }

    public void processPayment(String customer, double amount, boolean success) {
        if (success) {
            String msg = messages.getString("payment.success");
            System.out.println(MessageFormat.format(msg, customer, amount));
        } else {
            // BAD: Inline error
            System.out.println("Payment failed for " + customer + "! Amount: " + amount); // ❌

            // GOOD: Localized
            System.out.println(messages.getString("error.payment"));
        }
    }
}
