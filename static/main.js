window.addEventListener("load", () => {
  console.log("page is fully loaded");
  const zoneDisplayEl = document.getElementById("zone_display");
  const levelDisplayEl = document.getElementById("level_display");
  // const notesSectionEl = document.getElementById("notes_section");
  // const zoneSectionEl = document.getElementById("zone_section");
  // const zoneImageEl = document.getElementById("zone_image");
  // const zoneNotesEl = document.getElementById("zone_notes");
  // const buildSectionEl = document.getElementById("build_section");
  // const buildImageEl = document.getElementById("build_image");
  // const buildNotesEl = document.getElementById("build_notes");

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

    setTimeout(() => {
      handleUpdate();
    }, 500);
  };

  handleUpdate();
});
