const chatArea = document.getElementById("chat-area");
const chatForm = document.getElementById("chat-form");

let messages = [];

const API_URL = "http://localhost:8080/conversation";

chatForm.addEventListener("submit", async (e) => {
  e.preventDefault();

  const userInput = document.getElementById("user-input");
  const userMessage = userInput.value.trim();

  if (!userMessage) {
    return;
  }

  displayMessage("User", userMessage);
  showTypingIndicator(true);
  const response = await sendMessage(userMessage, messages);
  showTypingIndicator(false);
  messages = response.messages;
  displayMessage("Assistant", response.response);
  userInput.value = "";
});

function showTypingIndicator(show) {
  const typingIndicator = document.getElementById("typing-indicator");
  typingIndicator.style.display = show ? "block" : "none";
  const send = document.getElementById("send");
  send.style.display = show ? "none" : "block";
}

function displayMessage(role, content) {
  const messageElement = document.createElement("div");
  messageElement.classList.add("message");
  messageElement.innerHTML = `<strong>${role}:</strong> ${content}`;
  chatArea.appendChild(messageElement);

  // Scroll to the bottom of the chat area
  chatArea.scrollTop = chatArea.scrollHeight;
}

async function sendMessage(userMessage, messages) {
  try {
    const response = await fetch(API_URL, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ userMessage, messages }),
    });

    if (!response.ok) {
      throw new Error(`API request failed: ${response.statusText}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error(`Error while sending message: ${error.message}`);
  }
}
