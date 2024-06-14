-- Example PL/SQL script to demonstrate find_all_passwords and empty-string rules

-- Example 1: Finding all passwords
CREATE OR REPLACE PACKAGE find_passwords AS
  -- Declaration of passwords
  password1 VARCHAR2(100) := 'Password123!';
  password2 VARCHAR2(100) := 'Admin@456';
  password3 VARCHAR2(100) := 'UserPass789';

  -- Procedure to output passwords
  PROCEDURE output_passwords;
END find_passwords;
/

CREATE OR REPLACE PACKAGE BODY find_passwords AS
  PROCEDURE output_passwords IS
BEGIN
    -- Output passwords (for demonstration purposes)
    DBMS_OUTPUT.PUT_LINE('Password1: ' || password1);
    DBMS_OUTPUT.PUT_LINE('Password2: ' || password2);
    DBMS_OUTPUT.PUT_LINE('Password3: ' || password3);
END output_passwords;
END find_passwords;
/

-- Example 2: Using empty strings
CREATE OR REPLACE PACKAGE find_empty_string AS
  -- Declaration of empty strings
  empty_string1 VARCHAR2(100) := '';
  empty_string2 VARCHAR2(100);

  -- Procedure to output empty strings
  PROCEDURE output_empty_strings;
END find_empty_string;
/

CREATE OR REPLACE PACKAGE BODY find_empty_string AS
  PROCEDURE output_empty_strings IS
BEGIN
    -- Output empty strings (for demonstration purposes)
    DBMS_OUTPUT.
