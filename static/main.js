window.addEventListener("load", () => {
  console.log("page is fully loaded");
  const zoneDisplayEl = document.getElementById("zone_display");
  const levelDisplayEl = document.getElementById("level_display");
  const notesSectionEl = document.getElementById("notes_section");
  // const zoneSectionEl = document.getElementById("zone_section");
  // const zoneImageEl = document.getElementById("zone_image");
  // const zoneNotesEl = document.getElementById("zone_notes");
  // const buildSectionEl = document.getElementById("build_section");
  // const buildImageEl = document.getElementById("build_image");
  // const buildNotesEl = document.getElementById("build_notes");

  const allHelperSteps = [];

  const handleUpdate = async () => {
    let client = "stand_alone";

    const res = await fetch("/data", {
      method: "POST",
      body: JSON.stringify({ poe_client: client }),
    });
    const resJson = await res.json();

    const data = resJson[client];
    if (data) {
      zoneDisplayEl.textContent = data.zone;
      levelDisplayEl.textContent = data.level;
    }

    if (allHelperSteps.length > 0) {
      console.log("note", allHelperSteps[0].note);
      if (notesSectionEl.innerHTML !== allHelperSteps[0].note) {
        notesSectionEl.innerHTML = allHelperSteps[0].note;
      }
    }

    const codeTags = document.querySelectorAll("code");
    codeTags.forEach((codeTag) => {
      if (codeTag.getAttribute("role") !== "button") {
        const text = codeTag.textContent;
        codeTag.addEventListener("click", () => {
          navigator.clipboard.writeText(text);
        });
        codeTag.setAttribute("role", "button");
        codeTag.classList.add("clickToCopy");
      }
    });

    // @todo(nick-ng): change to websocket
    setTimeout(() => {
      handleUpdate();
    }, 500);
  };

  const loadHelpers = async () => {
    const res = await fetch("/helper");
    const resJson = await res.json();

    for (let i = 0; i < resJson.length; i++) {
      const res = await fetch(`/${resJson[i]}`);
      const resJson2 = await res.json();
      allHelperSteps.push(...resJson2);
    }

    handleUpdate();
  };

  loadHelpers();
});
