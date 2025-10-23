window.addEventListener("load", () => {
  const notesSectionEl = document.getElementById("notes_section");

  const handleUpdate = async () => {
    const paramsString = window.location.search;
    const searchParams = new URLSearchParams(paramsString);
    const client = searchParams.get("client") || "stand_alone";

    const res = await fetch("/data", {
      method: "POST",
      body: JSON.stringify({ poe_client: client }),
    });
    const resJson = await res.json();

    const data = resJson[client];
    if (data) {
      notesSectionEl.innerHTML = data.htmlNote;
    }

    // @todo(nick-ng): change to websocket
    setTimeout(() => {
      handleUpdate();
    }, 2000);
  };

  handleUpdate();
});
