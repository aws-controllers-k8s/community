window.addEventListener('DOMContentLoaded', (event) => {
  document.getElementById('mode').addEventListener('click', () => {
    document.body.classList.toggle('dark');
    localStorage.setItem('theme', document.body.classList.contains('dark') ? 'dark' : 'light');
  });

  let systemPreferenceDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  if (localStorage.getItem('theme') === 'light' || !systemPreferenceDark) {
    document.body.classList.remove('dark');
  }
});
