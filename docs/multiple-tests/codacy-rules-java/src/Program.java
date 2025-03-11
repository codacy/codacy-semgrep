import java.io.Console;
import de.hybris.platform.servicelayer.search.FlexibleSearchQuery;

class Program
    {
        public static void main(String[] args)
        {
            private static final String PASSWORD = "password" ; // Issue: Hardcoded password
            final FlexibleSearchQuery query = new FlexibleSearchQuery("SELECT {a.pk} FROM {TEST AS a} WHERE {a.uid} ="+ uid +" AND {a.visibleInAddressBook} = true");

            final FlexibleSearchQuery okquery = new FlexibleSearchQuery(
                "SELECT {a.pk} FROM {TEST AS a} WHERE {a.uid} = ?uid AND {a.visibleInAddressBook} = true"
            );
            okquery.addQueryParameter("uid", uid);
            System.out.println("This is a security risk: " + PASSWORD);
        }
    }

