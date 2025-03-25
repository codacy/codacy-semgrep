class myClass  
{
    public void MyMethod() 
    {
        int MyVariable = 10;
        bool anothervariable = true;
        int minLimit = 10 //this shouldn't be flagged
        // int maxLimit = 100; this shouldn't be flagged
    }
}
