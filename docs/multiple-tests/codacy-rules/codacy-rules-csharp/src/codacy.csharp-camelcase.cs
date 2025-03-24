class myClass  
{
    public void MyMethod() 
    {
        int MyVariable = 10;
        List<int> numberList = new List<int>();
        Dictionary<string, int> ageMap = new Dictionary<string, int>();
        MyClass instance = new MyClass();
        const int maxLimit = 100;

        // const int maxLimit = 100; this shouldn't be flagged
        // public void MyMethod() {}
    }
}
