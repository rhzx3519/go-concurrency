package contextpattern

import (
	"context"
	"fmt"
	"time"
)

// Context
func printGreeting(ctx context.Context) error {
	greeting, err := genGreeting(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func printFarewell(ctx context.Context) error {
	farewell, err := genFarewell(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", farewell)
	return nil
}

func genGreeting(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	switch locale, err := locale(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func genFarewell(ctx context.Context) (string, error) {
	switch locale, err := locale(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func locale(ctx context.Context) (string, error) {
	// Here we check to see whether our Context has provided a deadline. If it
	// did, and our system’s clock has advanced past the deadline, we simply
	// return with a special error defined in the context package,
	// DeadlineExceeded.
	if deadline, ok := ctx.Deadline(); ok {
		if deadline.Sub(time.Now().Add(1*time.Minute)) <= 0 {
			return "", context.DeadlineExceeded
		}
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(5 * time.Second):
	}
	return "EN/US", nil
}

// //////////////////////////////////////////////////////////////////////
// Context Value
//   - The key you use must satisfy Go’s notion of comparability, that is the
//     equality operators == and != need to return correct results when used.
//   - Values returned must be safe to access from multiple goroutines.
// Convention:
//   - First, they recommend you define a custom key-type in your package. As
//     long as other packages do the same, this prevents collisions within the
//     Context.
//   - Since the type you define for your package’s keys is unexported, other
//     packages cannot conflict with keys you generate within your package.
//     Since we don’t export the keys we use to store the data, we must therefore
//     export functions that retrieve the data for us.

type ctxKey int

const (
	ctxUserID ctxKey = iota
	ctxAuthToken
)

func UserID(c context.Context) string {
	return c.Value(ctxUserID).(string)
}
func AuthToken(c context.Context) string {
	return c.Value(ctxAuthToken).(string)
}

func ProcessRequest(userID, authToken string) {
	ctx := context.WithValue(context.Background(), ctxUserID, userID)
	ctx = context.WithValue(ctx, ctxAuthToken, authToken)
	HandleResponse(ctx)
}

func HandleResponse(ctx context.Context) {
	fmt.Printf("handling response for %v (auth: %v)", UserID(ctx), AuthToken(ctx))
}
