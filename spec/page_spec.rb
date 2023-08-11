require_relative "../src/configuration"
require_relative "../src/page"

RSpec.describe Page do
  subject(:page) { described_class.new(source_file: source_file, configuration: configuration) }

  let(:configuration) { Configuration.new(file_fixture("configuration/sample.yml")) }

  context "with frontmatter" do
    let(:source_file) { file_fixture("sample_pages/20230811_0001_full.md") }

    it { expect(page).to be_public }
    it { expect(page.uid).to eq("20230811_0001") }
    it { expect(page.short_uid).to eq("0001") }
    it { expect(page.slug).to eq("sample-slug") }
    it { expect(page.tags).to eq(["banana", "coconut"]) }
    it { expect(page.published_at).to eq(Date.parse("2023-08-11").to_time) }
    it { expect(page.body).to eq("<p>Sample content</p>\n") }
    it { expect(page.page_file).to eq("20230811_0001/sample-slug/index.html") }
    it { expect(page.redirect_file).to eq("20230811_0001/index.html") }
    it { expect(page.public_path).to eq("/20230811_0001/sample-slug") }
    it { expect(page.title).to eq("Page title") }
    it { expect(page).not_to be_hide_from_toc }
    it { expect(page.url).to eq("https://notes.musayev.com/20230811_0001/sample-slug") }
  end

  context "with no frontmatter" do
    let(:source_file) { file_fixture("sample_pages/20230811_0002_no-frontmatter.md") }

    it { expect(page).not_to be_public }
    it { expect(page.uid).to eq("20230811_0002") }
    it { expect(page.short_uid).to eq("0002") }
    it { expect(page.slug).to be_empty }
    it { expect(page.tags).to be_empty }
    it { expect(page.published_at).to eq(Date.parse("2023-08-11").to_time) }
    it { expect(page.body).to eq("<p>Sample content</p>\n") }
    it { expect(page.page_file).to eq("20230811_0002/index.html") }
    it { expect(page.redirect_file).to eq("20230811_0002/index.html") }
    it { expect(page.public_path).to eq("/20230811_0002/") }
    it { expect(page.title).to eq("20230811_0002") }
    it { expect(page).not_to be_hide_from_toc }
    it { expect(page.url).to eq("https://notes.musayev.com/20230811_0002/") }
  end

  context "when page is hidden from TOC" do
    let(:source_file) { file_fixture("sample_pages/20230811_0003_hidden.md") }

    it { expect(page).to be_hide_from_toc }
  end
end
