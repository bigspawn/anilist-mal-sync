// Package verniy is unofficial Anilist GraphQL API library.
//
// Verniy will generate graphql query string according to your request,
// make request to Anilist, parse response body, and convert it
// to struct. The goal of this library is to make a flexible and easy to
// call Anilist API.
//
//  // Init verniy.
//  v := verniy.New()
//
//  // Get anime One Piece data (with default fields).
//  data, err := v.GetAnime(21)
//
//  // Get anime One Piece data (with custom fields).
//  data, err := v.GetAnime(21,
//    verniy.MediaFieldID,
//    verniy.MediaFieldTitle(
//      verniy.MediaTitleFieldRomaji,
//      verniy.MediaTitleFieldEnglish,
//      verniy.MediaTitleFieldNative),
//    verniy.MediaFieldType,
//    verniy.MediaFieldFormat,
//    verniy.MediaFieldStatusV2,
//    verniy.MediaFieldDescription,
//    verniy.MediaFieldStartDate,
//    verniy.MediaFieldEndDate,
//    verniy.MediaFieldSeason,
//    verniy.MediaFieldSeasonYear)
//
// Functions
//
// There are alot of functions in verniy package but
// most of the time, you just need functions in `Client` struct.
// And it's recommended to use them to make your life easier.
// If you want to make a custom request, you can use function
// `MakeRequest()` (go to the function for more details).
//
// Parameters
//
// Yes, there are tons of custom types and functions but let
// your IDE auto-complete helps you choose the params for the
// functions. Not only constant value but also function (like
// `MediaFieldTitle` in the example) to fill the params.
// But, most of the functions have variadic parameters,
// so you don't need to actually fill it. There is already default
// value and you can read the code yourself to see the default
// value of the variadic parameters.
//
// If you want the complete list of available params,
// read the official docs (https://anilist.github.io/ApiV2-GraphQL-Docs/).
//
// Response
//
// Most of the functions in `Client` struct return a struct and error.
// Yes, most of the fields in the returned struct are pointers because,
// like any graphql, Anilist will not return fields that we didn't
// request. It is recommended to make your own pointer handler when
// using the field. For example.
//
//  // Handle *string to string.
//  func ptrToString(str *string) string {
//    if str == nil {
//      return ""
//    }
//    return *str
//  }
//
// Rate Limit
//
// Anilist has default rate limit 90 requests per minute. If you go over
// the rate limit you'll receive a 1-minute timeout. But, verniy has
// a built-in rate limiter to prevent the requests going over the limit.
// It will put your request on hold until the limit is available again.
package verniy
