using System.Threading.Tasks;
using Google.GenAI;
using Google.GenAI.Types;

public class GenerateContentSimpleText {
  public static async Task main() {
    // The client gets the API key from the environment variable `GEMINI_API_KEY`.
    var client = new Client();
    var response = await client.Models.GenerateContentAsync(
      model: "deepseek-v3.2", contents: "Explain how AI works in a few words"
    );
    var response2 = await client.Models.GenerateContentAsync(
      model: "gemini-2.5-flash", contents: "Explain how AI works in a few words"
    );
    Console.WriteLine(response.Candidates[0].Content.Parts[0].Text);
  }
}