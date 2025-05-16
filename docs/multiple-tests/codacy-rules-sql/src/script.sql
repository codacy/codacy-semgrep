GRANT ALL PRIVILEGES
ON mydb.*
TO 'myuser'@'%'
WITH GRANT OPTION;


GRANT ALL PRIVILEGES ON mydb.* TO myuser;




GRANT SELECT ON mydb.* TO scmdbi;

GRANT DELETE, INSERT, SELECT, UPDATE ON mydb.* TO scmdbi;

GRANT DELETE, INSERT, SELECT, UPDATE ON mydb.* TO scmd_dev_role;


SELECT fnd_profile.value('JTF_PROFILE_DEFAULT_NUM_ROWS') from dual;

SELECT * from table_name where lang = "US";
SELECT * from table_name where currency = "USD"; 
SELECT * from table_name where org_id = 123;

SELECT * FROM something WHERE lookup_type = 'ABC';
SELECT * FROM apps.fnd_lookup_values;

SELECT * FROM ap_all_invoices WHERE invoice_date > SYSDATE - 30;