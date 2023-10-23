package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	http.HandleFunc("/", handleMainRequest)
	port := ":8080"
	fmt.Printf("Server is running on port %s\n", port)
	http.ListenAndServe(port, nil)
}

func handleMainRequest(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	serverKey := os.Getenv("SERVER_KEY")
	midtransURL := "https://app.sandbox.midtrans.com/snap/v1/transactions"

	// Prepare the request payload
	payload := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     "order-csb-" + getCurrentTimestamp(),
			"gross_amount": 10000,
		},
		"credit_card": map[string]interface{}{
			"secure": true,
		},
		"customer_details": map[string]interface{}{
			"first_name": "Johny",
			"last_name":  "Kane",
			"email":      "testmidtrans@mailnesia.com",
			"phone":      "08111222333",
		},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", midtransURL, bytes.NewBuffer(requestBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(serverKey+":")))

	response, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	snapToken := result["token"].(string)
	htmlPage := getMainHtmlPage(snapToken)
	w.Write([]byte(htmlPage))
}

func getMainHtmlPage(snapToken string) string {
	return `
<html>
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://unpkg.com/spectre.css/dist/spectre.min.css">
    <link rel="stylesheet" href="https://unpkg.com/spectre.css/dist/spectre-exp.min.css">
    <link rel="stylesheet" href="https://unpkg.com/spectre.css/dist/spectre-icons.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/9.15.10/styles/monokai-sublime.min.css" />
    <style>pre{margin: -2em 0em -2em 0em;} .container{padding: 1em 1em;}</style>
  </head>

  <body class="container">
    <h4>Snap Payment Integration Demo</h4>
    <hr>
    <h5>Purchase Summary</h5>
    <small>
      <li><b>Customer Name:</b> Johny Kane</li>
      <li><b>Total Purchase:</b> IDR 10.000,-</li>
    </small>
    <br>
    <form id="snaphtml" onsubmit="return false" class="input-group">
      <input type="text" id="snap-token" value="` + snapToken + `" class="form-input">
      <button id="pay-button" class="btn btn-primary input-group-btn">Proceed to Payment</button>
    </form>
    <small>
      <ul>
        <li>Click "Proceed to Payment" to test Snap Popup. <a href="javascript:location.reload();">Refresh this iframe</a> to get a new Snap Token</li>
        <li>That string above is "Snap Transaction Token" retrieved from the API response, on the backend</li>
      </ul>
    </small>
    <br>
    <hr>
    <h4>1. Backend Implementation:</h4>
    <p>Call Midtrans Snap API to retrieve "Snap Transaction Token"</p>
    <pre>
      <code class="language-javascript">
// Your Go code here
      </code>
    </pre>
    <small>*Note: Your Go code replaces the Node.js code</small>
    <br><br><hr>

    <h4>2. Frontend Implementation</h4>
    <p>Pass "Snap Transaction Token" to frontend, import "snap.js" script tag, and call <code>snap.pay(&lt;snapToken&gt;)</code> to display the payment popup.</p>
    <pre>
      <code class="language-html" id="snapjs-view"></code>
    </pre>
        
    <div id="snapjs">
      <script type="text/javascript" src="https://app.sandbox.midtrans.com/snap/snap.js" data-client-key="YOUR_CLIENT_KEY"></script>

      <script type="text/javascript">
        var payButton = document.getElementById('pay-button');
        payButton.addEventListener('click', function() {
          var snapToken = document.getElementById('snap-token').value;
          snap.pay(snapToken);
        });
      </script>
    </div>
    <script>
      document.getElementById('snapjs-view').innerText = document.getElementById('snaphtml').innerHTML + document.getElementById('snapjs').innerHTML;
    </script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/9.15.10/highlight.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/9.15.10/languages/javascript.min.js"></script>
    <script>hljs.initHighlightingOnLoad();</script>
  </body>
</html>
`
}

func getCurrentTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}
