GRANT ALL PRIVILEGES
ON mydb.*
TO 'myuser'@'%'
WITH GRANT OPTION;


GRANT ALL PRIVILEGES ON mydb.* TO myuser;




GRANT SELECT ON mydb.* TO scmdbi;

GRANT DELETE, INSERT, SELECT, UPDATE ON mydb.* TO scmdbi;

GRANT DELETE, INSERT, SELECT, UPDATE ON mydb.* TO scmd_dev_role;


SELECT fnd_profile.value('JTF_PROFILE_DEFAULT_NUM_ROWS') from dual;