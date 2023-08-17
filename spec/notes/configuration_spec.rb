RSpec.describe Notes::Configuration do
  subject(:configuration) { described_class }

  let(:sample_site_path) { file_fixture_path.join("sample_site") }

  before { stub_env("NOTES_CONFIGURATION_PATH", expand("configuration.yml")) }

  it { expect(configuration.notes_path).to eq("/home/user/notes") }
  it { expect(configuration.build_path).to eq(expand("dist")) }
  it { expect(configuration.templates_path).to eq(expand("templates")) }
  it { expect(configuration.site_root_url).to eq("https://example.com") }
  it { expect(configuration.site_name).to eq("Site Name") }
  it { expect(configuration.author_name).to eq("Author Name") }
  it { expect(configuration.site_root_path).to eq("/") }
  it { expect(configuration.feed_url).to eq("https://example.com/feed.xml") }
  it { expect(configuration.feed_path).to eq("/feed.xml") }
  it { expect(configuration.assets_path).to eq(expand("images")) }
  it { expect(configuration.images_index_path).to eq(expand("images/index.json")) }

  def expand(rel_path)
    sample_site_path.join(rel_path).to_s
  end
end
