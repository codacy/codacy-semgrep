GRANT ALL PRIVILEGES
ON mydb.*
TO 'myuser'@'%'
WITH GRANT OPTION;


GRANT ALL PRIVILEGES ON mydb.* TO myuser;




GRANT SELECT ON mydb.* TO scmdbi;

GRANT DELETE, INSERT, SELECT, UPDATE ON mydb.* TO scmdbi;

GRANT DELETE, INSERT, SELECT, UPDATE ON mydb.* TO scmd_dev_role;


SELECT fnd_profile.value('JTF_PROFILE_DEFAULT_NUM_ROWS') from dual;

SELECT * FROM users WHERE language = 'US';
UPDATE accounts SET currency := 'USD';
UPDATE customers SET customers.org_id := '12345';

SELECT * FROM something WHERE lookup_type = 'ABC';
SELECT * FROM apps.fnd_lookup_values;

SELECT * FROM ap_all_invoices WHERE invoice_date > SYSDATE - 30;
SELECT * FROM RAC_tests

SELECT * FROM employees WHERE ROWNUM <= 10;