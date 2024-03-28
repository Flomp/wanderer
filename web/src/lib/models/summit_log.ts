
class SummitLog {
  id?: string;
  date: string;
  text?: string;

  constructor(date: string, params?: { id?: string, text?: string }) {
    this.date = date;
    this.id = params?.id;
    this.text = params?.text ?? "";
  }
}


export { SummitLog };
