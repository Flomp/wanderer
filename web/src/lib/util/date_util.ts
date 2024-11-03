export function isToday(date: Date) {
    const today = new Date();
    return date.setHours(0, 0, 0, 0) == today.setHours(0, 0, 0, 0)
}

export function dateExistsInList(targetDate: Date, dateList: Date[]): boolean {
    const targetYear = targetDate.getFullYear();
    const targetMonth = targetDate.getMonth();
    const targetDay = targetDate.getDate();

    for (const date of dateList) {
        const year = date.getFullYear();
        const month = date.getMonth();
        const day = date.getDate();

        if (year === targetYear && month === targetMonth && day === targetDay) {
            return true;
        }
    }

    return false;
}

export function isSameDay(d1: Date, d2: Date) {
    return d1.getFullYear() === d2.getFullYear() &&
        d1.getMonth() === d2.getMonth() &&
        d1.getDate() === d2.getDate();
}