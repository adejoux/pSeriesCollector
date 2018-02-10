import { SnmpcollectorPage } from './app.po';

describe('pseriescollector App', function() {
  let page: pseriescollectorPage;

  beforeEach(() => {
    page = new pseriescollectorPage();
  });

  it('should display message saying app works', () => {
    page.navigateTo();
    expect(page.getParagraphText()).toEqual('app works!');
  });
});
