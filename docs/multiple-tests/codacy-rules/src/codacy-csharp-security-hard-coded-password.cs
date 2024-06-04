using System;

namespace SecurityExample
{
    class Program
    {
        static void Main(string[] args)
        {
            var password = "password"; // Issue: Hardcoded password

            Console.WriteLine("This is a security risk: " + password);
        }
    }
}
