document.querySelectorAll("button").forEach((el) => {
  el.onclick = () => {
    alert("clicked")
  }
})

fetch("/api/file?filename=missing.txt")