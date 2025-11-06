public class TooTooLoop // Compliant - class name does not end with 'Exception'
{
  private string unusualCharacteristics;
  private bool appropriateForCommercialExploitation;
  // ...
}

public class RadioException: Exception // Compliant - correctly extends System.Exception
{
  public RadioException(string message, Exception inner): base(message, inner)
  {
     // ...
  }
}


public class RadioNLException : 
Exception // Compliant - correctly extends System.Exception
{
  public RadioNLException(string message, Exception inner): base(message, inner)
  {
     // ...
  }
}

public class TooTooLoopException // Noncompliant - this has nothing to do with Exception
{
  private string unusualCharacteristics;
  private bool appropriateForCommercialExploitation;
  // ...
}

public class GagaException // Noncompliant - does not derive from any Exception-based class
{
  public GagaException(string message, Exception inner)
  {
     // ...
  }
}