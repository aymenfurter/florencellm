const chatArea = document.getElementById("chat-area");
const chatForm = document.getElementById("chat-form");

let messages = [];

const API_URL = "http://localhost:8080/conversation";

// Event listener for chat form submission
chatForm.addEventListener("submit", async (e) => {
  e.preventDefault();

  const userInput = document.getElementById("user-input");
  const userMessage = userInput.value.trim();

  if (!userMessage) {
    return;
  }

  // Display user message in the chat area
  displayMessage("User", userMessage);

  // Show typing indicator
  showTypingIndicator(true);

  // Send message to the API
  const response = await sendMessage(userMessage, messages);

  // Hide typing indicator
  showTypingIndicator(false);

  // Update messages array with user and assistant messages
  messages = response.messages;

  // Display assistant message in the chat area
  displayMessage("Assistant", response.response);

  // Clear input field
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

const darkModeSwitch = document.getElementById("dark-mode-switch");

darkModeSwitch.addEventListener("change", () => {
  const html = document.documentElement;
  if (darkModeSwitch.checked) {
    html.setAttribute("data-theme", "dark");
    localStorage.setItem("theme", "dark");
  } else {
    html.removeAttribute("data-theme");
    localStorage.removeItem("theme");
  }
});

const storedTheme = localStorage.getItem("theme");
if (storedTheme) {
  document.documentElement.setAttribute("data-theme", storedTheme);
  darkModeSwitch.checked = storedTheme === "dark";
}

const startScratchButton = document.getElementById("start-scratch");
const userInput = document.getElementById("user-input");

startScratchButton.addEventListener("click", () => {
  location.href = location.href;
});
