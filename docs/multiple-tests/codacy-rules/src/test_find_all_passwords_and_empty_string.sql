-- Example PL/SQL script to demonstrate find_all_passwords and empty_string rules

-- Example 1: Finding all passwords
DECLARE
password1 VARCHAR2(100);
  password2 VARCHAR2(100);
  password3 VARCHAR2(100);
BEGIN
  -- Assigning passwords directly
  password1 := 'Password123!';
  password2 := 'Admin@456';
  password3 := 'UserPass789';

  -- Output the passwords (for demonstration purposes)
  DBMS_OUTPUT.PUT_LINE('Password1: ' || password1);
  DBMS_OUTPUT.PUT_LINE('Password2: ' || password2);
  DBMS_OUTPUT.PUT_LINE('Password3: ' || password3);
END;
/

-- Example 2: Using empty strings
DECLARE
empty_string1 VARCHAR2(100) := '';
  empty_string2 VARCHAR2(100);
BEGIN
  -- Assigning an empty string
  empty_string2 := '';

  -- Output the empty strings (for demonstration purposes)
  DBMS_OUTPUT.PUT_LINE('Empty String 1: ' || empty_string1);
  DBMS_OUTPUT.PUT_LINE('Empty String 2: ' || empty_string2);
END;
/
