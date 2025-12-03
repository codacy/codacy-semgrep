import java.util.Locale;
import java.util.ResourceBundle;

public class OrderApp {
    private static final ResourceBundle messages = ResourceBundle.getBundle("Messages", Locale.ENGLISH);

    public static void main(String[] args) {
        UILayer ui = new UILayer(messages);
        ui.showWelcome();

        OrderService orderService = new OrderService(messages);
        PaymentService paymentService = new PaymentService(messages);

        // GOOD: Using localization
        orderService.processOrder("Order123", 3);

        // BAD: Hardcoded UI label
        System.out.println("=== ORDER SUMMARY ==="); // ❌ should be localized

        // GOOD: Localized
        ui.showLabel("button.cancel");

        // Simulate payment success
        paymentService.processPayment("Carlos", 150.0, true);

        // Simulate payment failure
        paymentService.processPayment("Marie", 99.9, false);

        // BAD: Inline error
        System.out.println("Unable to generate invoice. Try again later."); // ❌
    }
}
