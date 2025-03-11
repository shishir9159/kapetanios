// Checking??? some basic staffs before a minor update
interface PrerequisitesCheck {
    state: "prerequisites"
    log: ""
}

interface UpdateVersionSelection {
    state: "updateVersionSelection"
    log: ""
}

interface Request {
    id: number;
    state: PrerequisitesCheck | UpdateVersionSelection;
}