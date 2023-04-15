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
