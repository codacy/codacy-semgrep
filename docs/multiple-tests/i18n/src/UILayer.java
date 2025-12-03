import java.util.ResourceBundle;

public class UILayer {
    private final ResourceBundle messages;

    public UILayer(ResourceBundle messages) {
        this.messages = messages;
    }

    public void showWelcome() {
        // BAD: Hardcoded welcome
        System.out.println("Welcome to Order Processing System"); // ❌

        // GOOD: Localized welcome
        System.out.println(messages.getString("app.start"));
    }

    public void showLabel(String key) {
        try {
            System.out.println(messages.getString(key));
        } catch (Exception e) {
            System.out.println("Missing i18n key: " + key); // ❌ fallback
        }
    }
}
