-- file: test_resource_injection.pls

DECLARE
resource1 VARCHAR2(100);
    resource2 VARCHAR2(100);
    resource3 VARCHAR2(100);
    resource4 VARCHAR2(100);
    resource5 VARCHAR2(100);
    resource6 VARCHAR2(100);
    resource7 VARCHAR2(100);
    resource8 VARCHAR2(100);
    resource9 VARCHAR2(100);
    resource10 VARCHAR2(100);
    resource11 VARCHAR2(100);
BEGIN
    resource1 := DBMS_CUBE.BUILD('arg1', 'arg2');
    resource2 := DBMS_FILE_TRANSFER.COPY_FILE('arg1', 'arg2', 'arg3');
    resource3 := DBMS_FILE_TRANSFER.GET_FILE('arg1', 'arg2', 'arg3');
    resource4 := DBMS_FILE_TRANSFER.PUT_FILE('arg1', 'arg2', 'arg3');
    resource5 := DBMS_SCHEDULER.GET_FILE('arg1', 'arg2', 'arg3');
    resource6 := DBMS_SCHEDULER.PUT_FILE('arg1', 'arg2', 'arg3');
    resource7 := DBMS_SCHEDULER.CREATE_PROGRAM('arg1', 'arg2', 'arg3');
    resource8 := DBMS_SERVICE.CREATE_SERVICE('arg1', 'arg2', 'arg3');
    resource9 := UTL_TCP.OPEN_CONNECTION('arg1', 'arg2');
    resource10 := UTL_SMTP.OPEN_CONNECTION('arg1', 'arg2');
    resource11 := WPG_DOCLOAD.DOWNLOAD_FILE('arg1', 'arg2');
END;
/
