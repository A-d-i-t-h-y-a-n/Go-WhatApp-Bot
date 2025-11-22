package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"context"

	"go.mau.fi/whatsmeow"
)

var qrCode string
var qrMutex sync.RWMutex
var pairingCode string
var isConnected bool
var connectionMutex sync.RWMutex
var client *whatsmeow.Client
var clientMutex sync.RWMutex

type QRResponse struct {
	QRCode      string `json:"qr_code"`
	PairingCode string `json:"pairing_code"`
	Connected   bool   `json:"connected"`
}

func SetQRCode(code string) {
	qrMutex.Lock()
	qrCode = code
	qrMutex.Unlock()
}

func SetConnected(connected bool) {
	connectionMutex.Lock()
	isConnected = connected
	connectionMutex.Unlock()
}

func SetClient(c *whatsmeow.Client) {
	clientMutex.Lock()
	client = c
	clientMutex.Unlock()
}

func StartServer() {
	http.HandleFunc("/qr", qrHandler)
	http.HandleFunc("/pairing", pairingHandler)
	http.HandleFunc("/status", statusHandler)

	fmt.Println("Web server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Web server error: %v\n", err)
	}
}

func qrHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	qrMutex.RLock()
	currentQR := qrCode
	qrMutex.RUnlock()

	connectionMutex.RLock()
	connected := isConnected
	connectionMutex.RUnlock()

	if connected {
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>WhatsApp Bot</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            background: #fafafa;
        }
        .container { 
            background: #fff;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            text-align: center;
        }
        .success { 
            color: #2d2d2d;
            font-size: 20px;
        }
        .success::before { 
            content: '✓';
            display: block;
            font-size: 48px;
            color: #4caf50;
            margin-bottom: 16px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success">Connected Successfully</div>
    </div>
</body>
</html>`)
		return
	}

	if currentQR == "" {
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>WhatsApp Bot</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            background: #fafafa;
        }
        .container { 
            background: #fff;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            text-align: center;
        }
        h2 { 
            font-size: 18px;
            font-weight: 500;
            color: #2d2d2d;
        }
    </style>
    <meta http-equiv="refresh" content="2">
</head>
<body>
    <div class="container">
        <h2>Waiting for QR Code...</h2>
    </div>
</body>
</html>`)
		return
	}

	qrDataURL := "https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=" + currentQR

	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>WhatsApp Bot - QR Code</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            background: #fafafa;
        }
        .container { 
            background: #fff;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            text-align: center;
        }
        h2 { 
            font-size: 18px;
            font-weight: 500;
            color: #2d2d2d;
            margin-bottom: 24px;
        }
        img { 
            border: 1px solid #e0e0e0;
            border-radius: 4px;
        }
        .info { 
            color: #666;
            margin-top: 24px;
            font-size: 14px;
        }
        a { 
            color: #1976d2;
            text-decoration: none;
            font-weight: 500;
        }
        a:hover { 
            text-decoration: underline;
        }
    </style>
    <meta http-equiv="refresh" content="5">
</head>
<body>
    <div class="container">
        <h2>Scan QR Code with WhatsApp</h2>
        <img src="%s" alt="QR Code">
        <div class="info">Or use pairing code</div>
        <a href="/pairing">Get Pairing Code</a>
    </div>
</body>
</html>`, qrDataURL)
}

func pairingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if r.Method == "POST" {
		connectionMutex.RLock()
		connected := isConnected
		connectionMutex.RUnlock()

		if connected {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"connected": true,
				"message":   "Already connected",
			})
			return
		}

		var req struct {
			PhoneNumber string `json:"phone_number"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid request",
			})
			return
		}

		clientMutex.RLock()
		currentClient := client
		clientMutex.RUnlock()

		if currentClient != nil {
			code, err := currentClient.PairPhone(ctx, req.PhoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			qrMutex.Lock()
			pairingCode = code
			qrMutex.Unlock()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":      true,
				"pairing_code": code,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Client not initialized",
		})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>WhatsApp Bot - Pairing Code</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            background: #fafafa;
        }
        .container { 
            background: #fff;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            text-align: center;
            min-width: 400px;
        }
        h2 { 
            font-size: 18px;
            font-weight: 500;
            color: #2d2d2d;
            margin-bottom: 8px;
        }
        .info { 
            color: #666;
            font-size: 14px;
            margin-bottom: 20px;
        }
        input { 
            padding: 10px 12px;
            width: 100%;
            font-size: 15px;
            border: 1px solid #d0d0d0;
            border-radius: 4px;
            margin-bottom: 16px;
        }
        input:focus { 
            outline: none;
            border-color: #1976d2;
        }
        button { 
            padding: 10px 24px;
            background: #1976d2;
            color: white;
            border: none;
            border-radius: 4px;
            font-size: 15px;
            cursor: pointer;
            font-weight: 500;
        }
        button:hover { 
            background: #1565c0;
        }
        .code { 
            font-size: 32px;
            font-weight: 600;
            color: #2d2d2d;
            margin: 24px 0;
            letter-spacing: 4px;
        }
        .error { 
            color: #d32f2f;
            margin-top: 16px;
        }
        a { 
            color: #1976d2;
            text-decoration: none;
            font-size: 14px;
        }
        a:hover { 
            text-decoration: underline;
        }
        .back { 
            margin-top: 24px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>Get Pairing Code</h2>
        <div class="info">Enter your phone number with country code</div>
        <input type="text" id="phone" placeholder="e.g., 1234567890" maxlength="15">
        <button onclick="getPairingCode()">Get Code</button>
        <div id="result"></div>
        <div class="back">
            <a href="/qr">Back to QR Code</a>
        </div>
    </div>
    <script>
        async function getPairingCode() {
            const phone = document.getElementById('phone').value.replace(/\D/g, '');
            if (!phone) {
                alert('Please enter a valid phone number');
                return;
            }
            
            const response = await fetch('/pairing', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ phone_number: phone })
            });
            
            const data = await response.json();
            const resultDiv = document.getElementById('result');
            
            if (data.success) {
                resultDiv.innerHTML = '<div class="code">' + data.pairing_code + '</div><div class="info">Enter this code in WhatsApp > Linked Devices > Link a Device</div>';
            } else {
                resultDiv.innerHTML = '<div class="error">Error: ' + data.error + '</div>';
            }
        }
    </script>
</body>
</html>`)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	qrMutex.RLock()
	currentQR := qrCode
	currentPairing := pairingCode
	qrMutex.RUnlock()

	connectionMutex.RLock()
	connected := isConnected
	connectionMutex.RUnlock()

	response := QRResponse{
		QRCode:      currentQR,
		PairingCode: currentPairing,
		Connected:   connected,
	}

	json.NewEncoder(w).Encode(response)
}