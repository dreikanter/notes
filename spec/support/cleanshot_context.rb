RSpec.shared_context "with cleanshot helpers" do
  let(:cleanshot_url) { "https://share.cleanshot.com/80085" }

  let(:cleanshot_direct_image_url) do
    "https://media.cleanshot.cloud/media/1334/itkYgfyMGKeaUT.jpeg?" \
    "response-content-disposition=attachment%3Bfilename%3DCleanShot" \
    "%25202023-08-06%2520at%252015.27.35.jpeg&Expires=1691367016"
  end

  let(:image_content) { file_fixture("banana.jpg").read }
  let(:cleanshot_original_file_name) { "CleanShot 2023-08-06 at 15.27.35.jpeg" }

  def stub_cleanshot_url
    stub_request(:get, "#{cleanshot_url}+")
  end

  def stub_cleanshot_direct_image_url
    stub_request(:get, cleanshot_direct_image_url)
  end
end
