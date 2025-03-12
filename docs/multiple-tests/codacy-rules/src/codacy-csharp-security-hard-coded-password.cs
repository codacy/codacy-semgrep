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

        public static bool? IsRegular(bool freqNoneOrNotPeriodic, bool freqPeriodical, IFrequency frequency)
        {
            if(frequency != null)
            {
                return false;
            }

  

            if (frequency.Days != 0)
                return false;

            if (frequency.Years == 1 && frequency.Months == 0)
                return null;

        }

        public static bool? IsRegular2(bool freqNoneOrNotPeriodic, bool freqPeriodical, IFrequency frequency)
        {
            if(frequency == null)
            {
                return false;
            }

  

            if (frequency.Days != 0)
                return false;

            if (frequency.Years == 1 && frequency.Months == 0)
                return null;

        }
    }
}
